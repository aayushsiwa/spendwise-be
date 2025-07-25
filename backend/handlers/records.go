package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"aayushsiwa/expense-tracker/db"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/secure"
)

func GetRecords(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT * FROM records ORDER BY date DESC")
	if err != nil {
		log.Println("Error querying records:", err)
		http.Error(w, "Failed to retrieve records", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var records []models.Record
	for rows.Next() {
		var rec models.Record
		err := rows.Scan(&rec.ID, &rec.Date, &rec.Description, &rec.Category, &rec.Amount, &rec.Type, &rec.Notes)
		if err != nil {
			log.Println("Error scanning row:", err)
			http.Error(w, "Failed to read records", http.StatusInternalServerError)
			return
		}

		// Optional: decrypt if storing encrypted data
		rec.Description, _ = secure.Decrypt(rec.Description)
		rec.Notes, _ = secure.Decrypt(rec.Notes)

		records = append(records, rec)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(records); err != nil {
		log.Println("Error encoding JSON:", err)
		http.Error(w, "Failed to encode records", http.StatusInternalServerError)
	}
}

func CreateRecord(w http.ResponseWriter, r *http.Request) {
	var rec models.Record
	if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
		log.Println("Invalid JSON input:", err)
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if rec.Date == "" || rec.Description == "" || rec.Category == "" || rec.Amount <= 0 || rec.Type == "" {
		http.Error(w, "Missing or invalid record fields", http.StatusBadRequest)
		return
	}

	// Optional: encrypt if storing encrypted data
	rec.Description, _ = secure.Encrypt(rec.Description)
	rec.Notes, _ = secure.Encrypt(rec.Notes)

	res, err := db.DB.Exec(`
		INSERT INTO records (date, description, category, amount, type, notes)
		VALUES (?, ?, ?, ?, ?, ?)`,
		rec.Date, rec.Description, rec.Category, rec.Amount, rec.Type, rec.Notes)
	if err != nil {
		log.Println("Error inserting record:", err)
		http.Error(w, "Failed to insert record", http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Println("Failed to retrieve last insert ID:", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	rec.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": "Created with ID - " + fmt.Sprint(rec.ID),
	}); err != nil {
		log.Println("Error encoding JSON:", err)
	}
}
