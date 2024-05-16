package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/mangkoyla/sub-api/converter"
)

var db *sql.DB

func Handler(w http.ResponseWriter, r *http.Request) {
	var err error
	db, err = sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		return
	}
	defer db.Close()

	format := r.URL.Query().Get("format")
	if format == "" {
		fmt.Println("Tolong masukkan parameter format")
		return
	}

	switch format {
	case "raw":
		handleRaw(w, r)
	case "clash":
		handleClash(w, r)
	default:
		http.Error(w, "Invalid format parameter", http.StatusBadRequest)
		return
	}
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
	rows, err := db.Query("SELECT * FROM proxies")
	if err != nil {
		fmt.Println("Error fetching accounts from database:", err)
		return nil
	}
	defer rows.Close()

	var accounts []interface{}
	for rows.Next() {
		var account interface{}
		err := rows.Scan(&account)
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
