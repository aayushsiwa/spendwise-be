package handlers

import (
	"encoding/csv"
	"net/http"
	"strconv"

	"aayushsiwa/expense-tracker/secure"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ExportCSV(c *gin.Context) {
	download := c.Query("download") == "true"

	rows, err := h.DB.Query("SELECT date, description, category_id, amount, type, note FROM records ORDER BY date DESC")
	if err != nil {
		c.String(http.StatusInternalServerError, "Error querying records")
		return
	}
	defer rows.Close()

	if download {
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=records.csv")
		writer := csv.NewWriter(c.Writer)
		_ = writer.Write([]string{"Date", "Description", "Category", "Amount", "Type", "Note"})

		for rows.Next() {
			var date, desc, cat, typ, note string
			var amt float64

			if err := rows.Scan(&date, &desc, &cat, &amt, &typ, &note); err != nil {
				continue
			}

			var categoryName string
			err := h.DB.QueryRow("SELECT name FROM categories WHERE id = ?", cat).Scan(&categoryName)
			if err != nil {
				c.String(http.StatusInternalServerError, "Error querying categories")
				return
			}

			decryptedDesc, _ := secure.Decrypt(desc)
			decryptedNote, _ := secure.Decrypt(note)

			record := []string{
				date,
				decryptedDesc,
				categoryName,
				strconv.FormatFloat(amt, 'f', 2, 64),
				typ,
				decryptedNote,
			}
			_ = writer.Write(record)
		}

		writer.Flush()
	} else {
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.Header("Content-Disposition", "inline; filename=records.txt")

		if _, err := c.Writer.Write([]byte("Date,Description,Category,Amount,Type,Note\n")); err != nil {
			c.String(http.StatusInternalServerError, "Error writing CSV header")
			return
		}

		for rows.Next() {
			var date, desc, cat, typ, note string
			var amt float64

			if err := rows.Scan(&date, &desc, &cat, &amt, &typ, &note); err != nil {
				continue
			}
			var categoryName string
			err := h.DB.QueryRow("SELECT name FROM categories WHERE id = ?", cat).Scan(&categoryName)
			if err != nil {
				c.String(http.StatusInternalServerError, "Error querying categories")
				return
			}

			decryptedDesc, _ := secure.Decrypt(desc)
			decryptedNote, _ := secure.Decrypt(note)

			line := date + "," + decryptedDesc + "," + categoryName + "," + strconv.FormatFloat(amt, 'f', 2, 64) + "," + typ + "," + decryptedNote + "\n"
			if _, err := c.Writer.Write([]byte(line)); err != nil {
				c.String(http.StatusInternalServerError, "Error writing CSV line")
				return
			}
		}
	}
}
