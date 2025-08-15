package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "config/config.yaml")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(os.Stderr, "Method: %s\n", r.Method)
		fmt.Fprintf(os.Stderr, "URL: %s\n", r.URL.String())
		fmt.Fprintf(os.Stderr, "Headers: %v\n", r.Header)
		fmt.Fprintf(os.Stderr, "---\n")
		w.WriteHeader(200)
	})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}