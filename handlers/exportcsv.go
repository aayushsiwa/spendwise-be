package handlers

import (
	"encoding/csv"
	"strconv"

	"aayushsiwa/expense-tracker/errors"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ExportCSV(c *gin.Context) {
	download := c.Query("download") == "true"

	records, err := h.Service.ExportRecords(c.Request.Context())
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	if download {
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=records.csv")
	} else {
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", "inline; filename=records.csv")
	}

	writer := csv.NewWriter(c.Writer)
	_ = writer.Write([]string{"Date", "Description", "Category", "Amount", "Type", "Note"})

	for _, rec := range records {
		if !download {
			line := rec.Date + "," + rec.Description + "," + rec.Category + "," +
				strconv.FormatFloat(rec.Amount, 'f', 2, 64) + "," +
				string(rec.Type) + "," + rec.Note + "\n"
			if _, err := c.Writer.Write([]byte(line)); err != nil {
				return
			}
			continue
		}

		_ = writer.Write([]string{
			rec.Date,
			rec.Description,
			rec.Category,
			strconv.FormatFloat(rec.Amount, 'f', 2, 64),
			string(rec.Type),
			rec.Note,
		})
	}

	writer.Flush()
}
