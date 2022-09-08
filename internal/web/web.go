package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"time"

	"ecksbee.com/telefacts-sec/internal/cache"
	"ecksbee.com/telefacts/pkg/hydratables"
	"ecksbee.com/telefacts/pkg/serializables"
	"github.com/gorilla/mux"
)

func SetupAndListen(ctx context.Context, wait time.Duration) {
	srv := setupServer()
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	listenForShutdown(ctx, wait, srv)
}

func setupServer() *http.Server {
	NewIdCache()
	r := newRouter()

	fmt.Println("telefacts-sec-browser<-0.0.0.0:8080")
	return &http.Server{
		Addr:         "0.0.0.0:8080",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}
}

func newRouter() http.Handler {
	appCache := cache.NewCache()
	dir, err := os.Getwd()
	if err != nil {
		dir = path.Join(".")
	}
	wd := os.Getenv("WD")
	if wd == "" {
		wd = dir
	}
	serializables.WorkingDirectoryPath = path.Join(wd, "wd")
	WorkingDirectoryPath = serializables.WorkingDirectoryPath
	gts := os.Getenv("GTS")
	if gts == "" {
		gts = dir
	}
	serializables.GlobalTaxonomySetPath = path.Join(gts, "gts")
	GlobalTaxonomySetPath = serializables.GlobalTaxonomySetPath
	hydratables.InjectCache(appCache)
	hydratables.HydrateEntityNames()
	hydratables.HydrateFundamentalSchema()
	hydratables.HydrateUnitTypeRegistry()
	r := mux.NewRouter()
	r.StrictSlash(true)
	Navigate(r)
	Render(r)
	conceptnetworkbrowser := http.FileServer(http.Dir((filepath.Join(wd, "wd", "goldlord-midas"))))
	r.PathPrefix("/browser").Handler(http.StripPrefix("/browser", conceptnetworkbrowser))
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		homeTmpl.Execute(w, nil)
	})
	return r
}

func listenForShutdown(ctx context.Context, grace time.Duration, srv *http.Server) {
	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println("Shutting down")
	ctx, cancel := context.WithTimeout(ctx, grace)
	defer cancel()
	srv.Shutdown(ctx)
	os.Exit(0)
}
