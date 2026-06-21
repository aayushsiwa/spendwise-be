package handlers

import (
	"encoding/csv"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ExportCSV(c *gin.Context) {
	download := c.Query("download") == "true"

	rows, err := h.DB.Query(`
		SELECT r.date, r.description, COALESCE(c.name, ''), r.amount, r.type, r.note
		FROM records r
		LEFT JOIN categories c ON r."categoryID" = c.id
		ORDER BY r.date DESC
	`)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error querying records")
		return
	}
	defer func() { _ = rows.Close() }()

	if download {
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=records.csv")
	} else {
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.Header("Content-Disposition", "inline; filename=records.txt")
	}

	writer := csv.NewWriter(c.Writer)
	_ = writer.Write([]string{"Date", "Description", "Category", "Amount", "Type", "Note"})

	for rows.Next() {
		var date, desc, category, typ, note string
		var amt float64

		if err := rows.Scan(&date, &desc, &category, &amt, &typ, &note); err != nil {
			continue
		}

		if !download {
			line := date + "," + desc + "," + category + "," + strconv.FormatFloat(amt, 'f', 2, 64) + "," + typ + "," + note + "\n"
			if _, err := c.Writer.Write([]byte(line)); err != nil {
				return
			}
			continue
		}

		_ = writer.Write([]string{
			date,
			desc,
			category,
			strconv.FormatFloat(amt, 'f', 2, 64),
			typ,
			note,
		})
	}

	writer.Flush()
}
