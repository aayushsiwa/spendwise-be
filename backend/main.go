package main

import (
	"log"
	"net/http"
	"os"

	"aayushsiwa/expense-tracker/db"
	"aayushsiwa/expense-tracker/handlers"
	"aayushsiwa/expense-tracker/secure"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	key := os.Getenv("ENCRYPTION_KEY")
	if len(key) != 32 {
		log.Fatalf("ENCRYPTION_KEY must be 32 bytes, got %d", len(key))
	}

	secure.SetKey([]byte(key))

}

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
