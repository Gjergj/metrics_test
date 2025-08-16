package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "config/config.yaml")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(os.Stderr, "Method: %s\n", r.Method)
		fmt.Fprintf(os.Stderr, "URL: %s\n", r.URL.String())
		fmt.Fprintf(os.Stderr, "Headers: %v\n", r.Header)

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading body: %v\n", err)
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		// Create filename with timestamp
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		filename := fmt.Sprintf("dumps/request_body_%s.txt", timestamp)

		// Create dumps directory if it doesn't exist
		err = os.MkdirAll("dumps", 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating dumps directory: %v\n", err)
			http.Error(w, "Error creating directory", http.StatusInternalServerError)
			return
		}

		// Write body to file
		err = os.WriteFile(filename, body, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing body to file: %v\n", err)
			http.Error(w, "Error saving request body", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(os.Stderr, "Body saved to: %s\n", filename)
		fmt.Fprintf(os.Stderr, "Body length: %d bytes\n", len(body))
		fmt.Fprintf(os.Stderr, "---\n")
		w.WriteHeader(200)
	})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
