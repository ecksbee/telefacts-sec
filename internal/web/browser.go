package web

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	neturl "net/url"
	"os"
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

var homeTmpl *template.Template
var searchTmpl *template.Template
var importTmpl *template.Template
var filingTmpl *template.Template
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
	idlock    sync.RWMutex
	idonce    sync.Once
	idCache   *gocache.Cache
	pathCache *gocache.Cache
)

func NewIdCache() *gocache.Cache {
	idonce.Do(func() {
		idCache = gocache.New(gocache.NoExpiration, gocache.NoExpiration)
		pathCache = gocache.New(gocache.NoExpiration, gocache.NoExpiration)
	})
	return idCache
}

func init() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	filingTmpl, err = template.ParseFiles(path.Join(dir, "filing.tmpl"))
	if err != nil {
		panic(err)
	}
	homeTmpl, err = template.ParseFiles(path.Join(dir, "home.tmpl"))
	if err != nil {
		panic(err)
	}
	searchTmpl, err = template.ParseFiles(path.Join(dir, "search.tmpl"))
	if err != nil {
		panic(err)
	}
	importTmpl, err = template.ParseFiles(path.Join(dir, "import.tmpl"))
	if err != nil {
		panic(err)
	}
	throttle.StartSECThrottle()
}

func Navigate(r *mux.Router) {
	r.Path("/review/{hash}").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Error: incorrect verb, "+r.Method, http.StatusInternalServerError)
			return
		}
		vars := mux.Vars(r)
		hash := vars["hash"]
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		idlock.RLock()
		mypath := "404"
		idlock.RLock()
		if path, found := pathCache.Get(hash); found {
			mypath = path.(string)
		}
		idlock.RUnlock()
		data := map[string]string{
			"Hash": hash,
			"Path": mypath,
		}
		filingTmpl.Execute(w, data)
	})
	r.Path("/{cik}/{filingid}/hash").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, "Error: incorrect verb, "+r.Method, http.StatusInternalServerError)
			return
		}
		<-time.After(2 * time.Second)
		idlock.RLock()
		cachedid := ""
		vars := mux.Vars(r)
		cik := vars["cik"]
		filingid := vars["filingid"]
		cachekey := cik + "/" + filingid
		idlock.RLock()
		if x, found := idCache.Get(cachekey); found {
			cachedid = x.(string)
			data, err := json.Marshal(map[string]string{
				"hash": cachedid,
			})
			if err != nil {
				http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
			idlock.RUnlock()
			return
		}
		idlock.RUnlock()
		w.WriteHeader(http.StatusNotFound)
	})
	r.Path("/import").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		trunc := path.Dir(mypath)
		filingid := path.Base(trunc)
		cik := path.Base(path.Dir(trunc))
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		idlock.RLock()
		cachedid := ""
		if x, found := idCache.Get(cik + "/" + filingid); found {
			cachedid = x.(string)
		}
		idlock.RUnlock()
		data := map[string]string{
			"AccessionNumber": filingid,
			"Cik":             cik,
		}
		importTmpl.Execute(w, data)
		go func() {
			if cachedid != "" {
				return
			} else {
				id, err := serializables.Download(
					"https://www.sec.gov/Archives/edgar/data/"+cik+"/"+filingid,
					WorkingDirectoryPath, GlobalTaxonomySetPath, throttle.Throttle)
				if err != nil {
					fmt.Printf("%v", err)
					return
				}
				idlock.Lock()
				defer idlock.Unlock()
				idCache.Set(cik+"/"+filingid, id, gocache.DefaultExpiration)
				pathCache.Set(id, mypath, gocache.DefaultExpiration)
			}
		}()
	})
	r.Path("/company-search").Methods("POST").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Error: incorrect verb, "+r.Method, http.StatusInternalServerError)
			return
		}
		r.ParseForm()
		search := r.Form["company-name"]
		typeid := r.Form["form-type"]
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
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
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
		searchTmpl.Execute(w, data)
	})
}
