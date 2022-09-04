package web

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	neturl "net/url"
	"os"
	"path"
	"strconv"
	"time"

	"ecksbee.com/telefacts-sec/internal/actions"
	"ecksbee.com/telefacts-sec/pkg/throttle"
	"github.com/gorilla/mux"
	"golang.org/x/net/html/charset"
)

var homeTmpl *template.Template
var searchTmpl *template.Template
var importTmpl *template.Template
var filingTmpl *template.Template

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

func init() {
	dir, err := os.Getwd()
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
	r.Path("/{type}/{cik}/{filingid}").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Error: incorrect verb, "+r.Method, http.StatusInternalServerError)
			return
		}
		vars := mux.Vars(r)
		typeid := vars["type"]
		cik := vars["cik"]
		filingid := vars["filingid"]
		switch typeid {
		case "8k", "10k", "10q", "485bpos":
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/html")
			data := map[string]string{
				"Type":     typeid,
				"Cik":      cik,
				"Filingid": filingid,
			}
			filingTmpl.Execute(w, data)
		default:
			http.Error(w, "Error: invalid type '"+typeid+"'", http.StatusBadRequest)
		}
	})
	r.Path("/{cik}/{filingid}").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Error: incorrect verb, "+r.Method, http.StatusInternalServerError)
			return
		}
		vars := mux.Vars(r)
		filingid := vars["filingid"]
		cik := vars["cik"]
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		data := map[string]string{
			"AccessionNumber": filingid,
			"Cik":             cik,
		}
		importTmpl.Execute(w, data)
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
		fmt.Printf("%s", string(body))
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
