package web

import (
	"net/http"
	neturl "net/url"

	"ecksbee.com/telefacts-sec/internal/cache"
	"github.com/gorilla/mux"
)

func Catalog() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Error: incorrect verb, "+r.Method, http.StatusInternalServerError)
			return
		}
		vars := mux.Vars(r)
		id := vars["id"]
		if len(id) <= 0 {
			http.Error(w, "Error: invalid id '"+id+"'", http.StatusBadRequest)
			return
		}
		data, err := cache.MarshalCatalog(id)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

func Renderable() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Error: incorrect verb, "+r.Method, http.StatusInternalServerError)
			return
		}
		vars := mux.Vars(r)
		id := vars["id"]
		if len(id) <= 0 {
			http.Error(w, "Error: invalid id '"+id+"'", http.StatusBadRequest)
			return
		}
		hash := vars["hash"]
		if len(hash) <= 0 {
			http.Error(w, "Error: invalid roote", http.StatusBadRequest)
			return
		}

		data, err := cache.MarshalRenderable(id, hash)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

func Expressable() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Error: incorrect verb, "+r.Method, http.StatusInternalServerError)
			return
		}
		vars := mux.Vars(r)
		id := vars["id"]
		if len(id) <= 0 {
			http.Error(w, "Error: invalid id '"+id+"'", http.StatusBadRequest)
			return
		}
		parsedquery, err := neturl.ParseQuery(r.URL.RawQuery)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		name, err := neturl.QueryUnescape(parsedquery.Get("name"))
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
			return
		}
		contextref, err := neturl.QueryUnescape(parsedquery.Get("contextref"))
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
			return
		}
		data, err := cache.MarshalExpressable(id, name, contextref)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

func Render(r *mux.Router) {
	foldersRoute := r.PathPrefix("/browser/folders").Subrouter()
	foldersRoute.HandleFunc("/{id}", Catalog()).Methods("GET")
	projectIDRoute := foldersRoute.PathPrefix("/{id}").Subrouter()
	projectIDRoute.HandleFunc("/facts", Expressable()).Methods("GET")
	projectIDRoute.HandleFunc("/{hash}", Renderable()).Methods("GET")
}
