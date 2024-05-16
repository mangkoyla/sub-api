package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/mangkoyla/sub-api/converter"
	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		return
	}
	defer db.Close()

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
	accounts := fetchAccountsFromDB()
	params := r.URL.Query()
	output := converter.ToRaw(accounts, params)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))
}

func handleClash(w http.ResponseWriter, r *http.Request) {
	accounts := fetchAccountsFromDB()
	params := r.URL.Query()
	output := converter.ToClash(accounts, params)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))
}

func fetchAccountsFromDB() []interface{} {
	rows, err := db.Query("SELECT * FROM proxies") // Sesuaikan dengan query yang sesuai
	if err != nil {
		fmt.Println("Error fetching accounts from database:", err)
		return nil
	}
	defer rows.Close()

	var accounts []interface{}
	for rows.Next() {
		var account interface{} // Ganti dengan tipe data yang sesuai dengan struktur akun Anda
		err := rows.Scan(&account) // Sesuaikan dengan struktur akun Anda
		if err != nil {
			fmt.Println("Error scanning account from database:", err)
			continue
		}
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		fmt.Println("Error iterating over accounts:", err)
		return nil
	}
	return accounts
}
