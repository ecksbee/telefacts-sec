package web

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"net/http"
	neturl "net/url"
	"path"
	"strconv"
	"sync"
	"time"

	"ecksbee.com/telefacts-sec/internal/actions"
	"ecksbee.com/telefacts-sec/pkg/serializables"
	"ecksbee.com/telefacts-sec/pkg/throttle"
	"github.com/gorilla/mux"
	gocache "github.com/patrickmn/go-cache"
	"golang.org/x/net/html/charset"
)

var WorkingDirectoryPath string
var GlobalTaxonomySetPath string

type EdgarRss struct {
	XMLName  xml.Name   `xml:"feed"`
	XMLAttrs []xml.Attr `xml:",any,attr"`
	Updated  []struct {
		XMLName  xml.Name
		XMLAttrs []xml.Attr `xml:",any,attr"`
		CharData string     `xml:",chardata"`
	} `xml:"updated"`
	Entry []struct {
		XMLName  xml.Name
		XMLAttrs []xml.Attr `xml:",any,attr"`
		Title    []struct {
			XMLName  xml.Name
			XMLAttrs []xml.Attr `xml:",any,attr"`
			CharData string     `xml:",chardata"`
		} `xml:"title"`
		Link []struct {
			XMLName  xml.Name
			XMLAttrs []xml.Attr `xml:",any,attr"`
		} `xml:"link"`
		Summary []struct {
			XMLName  xml.Name
			XMLAttrs []xml.Attr `xml:",any,attr"`
			CharData string     `xml:",chardata"`
		} `xml:"summary"`
		Updated []struct {
			XMLName  xml.Name
			XMLAttrs []xml.Attr `xml:",any,attr"`
			CharData string     `xml:",chardata"`
		} `xml:"updated"`
		Category []struct {
			XMLName  xml.Name
			XMLAttrs []xml.Attr `xml:",any,attr"`
			CharData string     `xml:",chardata"`
		} `xml:"category"`
		Id []struct {
			XMLName  xml.Name
			XMLAttrs []xml.Attr `xml:",any,attr"`
			CharData string     `xml:",chardata"`
		} `xml:"id"`
	} `xml:"entry"`
}

type SearchResult struct {
	SearchText string
	Results    []struct {
		Title                  string
		Summary                string
		EdgarUrl               string
		PercentEncodedEdgarUrl string
	}
}

var (
	idlock  sync.RWMutex
	idonce  sync.Once
	idCache *gocache.Cache
)

func NewIdCache() *gocache.Cache {
	idonce.Do(func() {
		idCache = gocache.New(gocache.NoExpiration, gocache.NoExpiration)
	})
	return idCache
}

func init() {
	throttle.StartSECThrottle()
}

func ServeApi(r *mux.Router) {
	r.Path("/hash").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Error: incorrect verb, "+r.Method, http.StatusInternalServerError)
			return
		}
		parsedquery, err := neturl.ParseQuery(r.URL.RawQuery)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		mypath, err := neturl.QueryUnescape(parsedquery.Get("path"))
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if mypath == "" {
			http.Error(w, "no path parameter", http.StatusBadRequest)
			return
		}
		trunc := path.Dir(mypath)
		filingid := path.Base(trunc)
		cik := path.Base(path.Dir(trunc))
		cachedid := ""
		cachekey := cik + "/" + filingid
		idlock.RLock()
		if x, found := idCache.Get(cachekey); found {
			cachedid = x.(string)
			data, err := json.Marshal(map[string]string{
				"hash": cachedid,
			})
			if err != nil {
				http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
				idlock.RUnlock()
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
			idlock.RUnlock()
			return
		}
		idlock.RUnlock()
		id, err := serializables.Download(
			"https://www.sec.gov/Archives/edgar/data/"+cik+"/"+filingid,
			WorkingDirectoryPath, GlobalTaxonomySetPath, throttle.Throttle)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		idlock.Lock()
		defer idlock.Unlock()
		idCache.Set(cik+"/"+filingid, id, gocache.DefaultExpiration)
		data, err := json.Marshal(map[string]string{
			"hash": id,
		})
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})
	r.Path("/company-search").Methods("POST").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Error: incorrect verb, "+r.Method, http.StatusInternalServerError)
			return
		}
		r.ParseMultipartForm(1024 * 8)
		search := r.PostForm["company-name"]
		typeid := r.PostForm["form-type"]
		if len(search) <= 0 || len(typeid) <= 0 {
			http.Error(w, "Missing parameters, company-name or form-type", http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/plaint")
			return
		}
		y := time.Now().Year()
		edgarSearchText := `company-name%3D"` + neturl.QueryEscape(search[0]) + `"%20AND%20form-type%3D%28` + neturl.QueryEscape(typeid[0]) + `%2A%29`
		queryText := `text=` + edgarSearchText + `&start=1&count=80&first=` + strconv.Itoa(y) + `&last=1994&output=atom`
		finalUrl := neturl.URL{
			Scheme:   `https`,
			Host:     `www.sec.gov`,
			Path:     `cgi-bin/srch-edgar`,
			RawQuery: queryText,
		}
		body, err := actions.Scrape(finalUrl.String(), throttle.Throttle)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		reader := bytes.NewReader(body)
		decoder := xml.NewDecoder(reader)
		decoder.CharsetReader = charset.NewReaderLabel
		rss := EdgarRss{}
		err = decoder.Decode(&rss)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		data := SearchResult{
			SearchText: search[0],
			Results: make([]struct {
				Title                  string
				Summary                string
				EdgarUrl               string
				PercentEncodedEdgarUrl string
			}, 0),
		}
		for _, entry := range rss.Entry {
			edgarUrl := ""
			for _, link := range entry.Link {
				for _, linkAttr := range link.XMLAttrs {
					if linkAttr.Name.Local == "href" {
						edgarUrl = linkAttr.Value
						break
					}
				}
			}
			if edgarUrl == "" {
				continue
			}
			data.Results = append(data.Results, struct {
				Title                  string
				Summary                string
				EdgarUrl               string
				PercentEncodedEdgarUrl string
			}{
				Title:                  entry.Title[0].CharData,
				Summary:                entry.Summary[0].CharData,
				EdgarUrl:               edgarUrl,
				PercentEncodedEdgarUrl: neturl.QueryEscape(edgarUrl),
			})
		}
		ret, err := json.Marshal(data)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			idlock.RUnlock()
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(ret)
	})
}
