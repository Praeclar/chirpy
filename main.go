package main

import (
	"net/http"
)

func main() {
	serveMux := http.NewServeMux()

	server := http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}

	serveMux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))

	handler := func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(200)
		writer.Write([]byte("OK"))
	}

	serveMux.HandleFunc("/healthz", handler)

	server.ListenAndServe()
}
