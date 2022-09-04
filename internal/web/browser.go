package web

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"

	"ecksbee.com/telefacts-sec/internal/actions"
	"ecksbee.com/telefacts-sec/pkg/throttle"
	"github.com/gorilla/mux"
	"golang.org/x/net/html/charset"
)

var homeTmpl *template.Template
var searchTmpl *template.Template
var cikTmpl *template.Template
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
	cikTmpl, err = template.ParseFiles(path.Join(dir, "cik.tmpl"))
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
	r.Path("/{type}/{cik}").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Error: incorrect verb, "+r.Method, http.StatusInternalServerError)
			return
		}
		vars := mux.Vars(r)
		typeid := vars["type"]
		cik := vars["cik"]
		switch typeid {
		case "8k", "10k", "10q", "485bpos":
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/html")
			data := map[string]string{
				"Type": typeid,
				"Cik":  cik,
			}
			cikTmpl.Execute(w, data)
		default:
			http.Error(w, "Error: invalid type '"+typeid+"'", http.StatusBadRequest)
		}
	})
	r.Path("/company-search").Methods("POST").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Error: incorrect verb, "+r.Method, http.StatusInternalServerError)
			return
		}
		r.ParseForm()
		search := r.Form["company-name"]
		typeid := r.Form["form-type"]
		url := `https://www.sec.gov/cgi-bin/srch-edgar?text=COMPANY-NAME%3D%22` + search[0] + `%22%20AND%20form-type%3D%28` + typeid[0] + `%2A%29&start=1&count=80&first=2022&last=2022&output=atom`
		body, err := actions.Scrape(url, throttle.Throttle)
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
		data := map[string]string{
			"SearchText": search[0],
			"Results":    string(body),
		}
		searchTmpl.Execute(w, data)
	})
}
