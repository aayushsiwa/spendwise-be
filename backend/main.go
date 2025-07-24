package main

import (
	"log"
	"net/http"

	"aayushsiwa/expense-tracker/db"
	"aayushsiwa/expense-tracker/handlers"
)

func main() {
	db.Init("records.db")

	http.HandleFunc("/api/records", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			handlers.GetRecords(w, r)
		} else if r.Method == "POST" {
			handlers.CreateRecord(w, r)
		}
	})

	http.HandleFunc("/api/export/csv", handlers.ExportCSV)

	log.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
