package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}

func main() {
	serveMux := http.NewServeMux()
	cfg := apiConfig{}

	server := http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}

	serveMux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	serveMux.HandleFunc("/healthz", getHealthzHandler)

	serveMux.HandleFunc("/metrics", func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(200)

		countString := fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())

		writer.Write([]byte(countString))
	})

	serveMux.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Store(0)

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)

	})

	server.ListenAndServe()
}

func getHealthzHandler(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}
