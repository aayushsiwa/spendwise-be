package handlers

import (
	"encoding/csv"
	"net/http"
	"strconv"

	"aayushsiwa/expense-tracker/db"
	"aayushsiwa/expense-tracker/secure"
)

func ExportCSV(w http.ResponseWriter, r *http.Request) {
	rows, _ := db.DB.Query("SELECT date, description, category, amount, type, notes FROM records ORDER BY date DESC")
	defer rows.Close()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=records.csv")

	writer := csv.NewWriter(w)
	writer.Write([]string{"Date", "Description", "Category", "Amount", "Type", "Notes"})

	for rows.Next() {
		var date, desc, cat, typ, notes string
		var amt float64
		rows.Scan(&date, &desc, &cat, &amt, &typ, &notes)

		desc, _ = secure.Decrypt(desc)
		notes, _ = secure.Decrypt(notes)

		writer.Write([]string{date, desc, cat, strconv.FormatFloat(amt, 'f', 2, 64), typ, notes})
	}

	writer.Flush()
}
