package web

import (
	"encoding/json"
	"net/http"
	neturl "net/url"
	"path"
	"strconv"
	"sync"
	"time"

	"ecksbee.com/telefacts-sec/pkg/serializables"
	"ecksbee.com/telefacts-sec/pkg/throttle"
	"github.com/gorilla/mux"
	gocache "github.com/patrickmn/go-cache"
)

var WorkingDirectoryPath string
var GlobalTaxonomySetPath string

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
		yearVal := r.PostForm["year"]
		if len(search) <= 0 || len(typeid) <= 0 || len(yearVal) <= 0 {
			http.Error(w, "Missing parameters: company-name, year, or form-type", http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/plain")
			return
		}
		year, err := strconv.Atoi(yearVal[0])
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if year > time.Now().Year() {
			http.Error(w, "Invalid year value", http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/plain")
			return
		}
		data, err := getFilings(search[0], typeid[0], yearVal[0])
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
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
