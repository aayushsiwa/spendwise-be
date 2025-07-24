package handlers

import (
	"encoding/json"
	"net/http"

	"aayushsiwa/expense-tracker/db"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/secure"
)

func GetRecords(w http.ResponseWriter, r *http.Request) {
	rows, _ := db.DB.Query("SELECT id, date, description, category, amount, type, notes FROM records ORDER BY date DESC")
	defer rows.Close()

	var records []models.Record
	for rows.Next() {
		var rec models.Record
		rows.Scan(&rec.ID, &rec.Date, &rec.Description, &rec.Category, &rec.Amount, &rec.Type, &rec.Notes)

		rec.Description, _ = secure.Decrypt(rec.Description)
		rec.Notes, _ = secure.Decrypt(rec.Notes)

		records = append(records, rec)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

func CreateRecord(w http.ResponseWriter, r *http.Request) {
	var rec models.Record
	json.NewDecoder(r.Body).Decode(&rec)

	rec.Description, _ = secure.Encrypt(rec.Description)
	rec.Notes, _ = secure.Encrypt(rec.Notes)

	res, _ := db.DB.Exec(`INSERT INTO records (date, description, category, amount, type, notes)
                          VALUES (?, ?, ?, ?, ?, ?)`,
		rec.Date, rec.Description, rec.Category, rec.Amount, rec.Type, rec.Notes)

	id, _ := res.LastInsertId()
	rec.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rec)
}
