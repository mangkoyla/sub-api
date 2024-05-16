package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/mangkoyla/sub-api/converter"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			format := r.URL.Query().Get("format")
			switch format {
			case "raw":
				handleRaw(w, r)
			case "clash":
				handleClash(w, r)
			default:
				http.Error(w, "Invalid format parameter", http.StatusBadRequest)
				return
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
	})

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "3000"
	}
	fmt.Printf("Server running at http://localhost:%s/\n", PORT)
	http.ListenAndServe(":"+PORT, nil)
}

func handleRaw(w http.ResponseWriter, r *http.Request) {
	accounts := []interface{}{}
	params := r.URL.Query()
	output := converter.ToRaw(accounts, params)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))
}

func handleClash(w http.ResponseWriter, r *http.Request) {
	accounts := []interface{}{}
	params := r.URL.Query()
	output := converter.ToClash(accounts, params)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))
}
