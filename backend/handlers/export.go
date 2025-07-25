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
	download := c.Query("download") == "true"

	rows, err := db.DB.Query("SELECT date, description, category, amount, type, notes FROM records ORDER BY date DESC")
	if err != nil {
		c.String(http.StatusInternalServerError, "Error querying records")
		return
	}
	defer rows.Close()

	if download {
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=records.csv")
		writer := csv.NewWriter(c.Writer)
		_ = writer.Write([]string{"Date", "Description", "Category", "Amount", "Type", "Notes"})

		for rows.Next() {
			var date, desc, cat, typ, notes string
			var amt float64

			if err := rows.Scan(&date, &desc, &cat, &amt, &typ, &notes); err != nil {
				continue
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
	} else {
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.Header("Content-Disposition", "inline; filename=records.txt")

		c.Writer.Write([]byte("Date,Description,Category,Amount,Type,Notes\n"))

		for rows.Next() {
			var date, desc, cat, typ, notes string
			var amt float64

			if err := rows.Scan(&date, &desc, &cat, &amt, &typ, &notes); err != nil {
				continue
			}

			decryptedDesc, _ := secure.Decrypt(desc)
			decryptedNotes, _ := secure.Decrypt(notes)

			line := date + "," + decryptedDesc + "," + cat + "," + strconv.FormatFloat(amt, 'f', 2, 64) + "," + typ + "," + decryptedNotes + "\n"
			c.Writer.Write([]byte(line))
		}
	}
}
