package handlers

import (
	"encoding/csv"
	"net/http"
	"strconv"

	"aayushsiwa/expense-tracker/db"
	"aayushsiwa/expense-tracker/secure"

	"github.com/gin-gonic/gin"
)

func ExportCSV(c *gin.Context) {
	rows, err := db.DB.Query("SELECT date, description, category, amount, type, notes FROM records ORDER BY date DESC")
	if err != nil {
		c.String(http.StatusInternalServerError, "Error querying records")
		return
	}
	defer rows.Close()

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=records.csv")

	writer := csv.NewWriter(c.Writer)
	_ = writer.Write([]string{"Date", "Description", "Category", "Amount", "Type", "Notes"})

	for rows.Next() {
		var date, desc, cat, typ, notes string
		var amt float64

		if err := rows.Scan(&date, &desc, &cat, &amt, &typ, &notes); err != nil {
			continue // or log the error and break
		}

		decryptedDesc, _ := secure.Decrypt(desc)
		decryptedNotes, _ := secure.Decrypt(notes)

		record := []string{
			date,
			decryptedDesc,
			cat,
			strconv.FormatFloat(amt, 'f', 2, 64),
			typ,
			decryptedNotes,
		}
		_ = writer.Write(record)
	}

	writer.Flush()
}
