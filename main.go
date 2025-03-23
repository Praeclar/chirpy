package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

	serveMux.HandleFunc("GET /api/healthz", getHealthzHandler)

	serveMux.HandleFunc("GET /admin/metrics", func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.WriteHeader(200)

		countString := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())

		writer.Write([]byte(countString))
	})

	serveMux.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Store(0)

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)

	})

	serveMux.HandleFunc("POST /api/validate_chirp", getValidateHandler)

	server.ListenAndServe()
}

func getValidateHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type errorOutput struct {
		Error string `json:"error"`
	}
	type validOutput struct {
		// Valid bool `json:"valid"`
		Cleaned_body string `json:"cleaned_body"`
	}

	errorAny := errorOutput{
		Error: "Something went wrong",
	}
	errAny, _ := json.Marshal(errorAny)
	errorTooLong := errorOutput{
		Error: "Chirp is too long",
	}
	errLong, _ := json.Marshal(errorTooLong)

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		w.WriteHeader(500)
		w.Write(errAny)
		return
	}

	if len(params.Body) > 140 {
		w.WriteHeader(400)
		w.Write(errLong)
		return
	}

	splitBody := strings.Split(params.Body, " ")

	for i, part := range splitBody {
		lwrdStr := strings.ToLower(part)
		if lwrdStr == "kerfuffle" || lwrdStr == "sharbert" || lwrdStr == "fornax" {
			splitBody[i] = "****"
		}
	}

	validVals := validOutput{
		Cleaned_body: strings.Join(splitBody, " "),
	}

	dat, err := json.Marshal(validVals)
	if err != nil {
		w.WriteHeader(500)
		w.Write(errAny)
		return
	}

	w.WriteHeader(200)
	w.Write(dat)
}

func getHealthzHandler(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}
