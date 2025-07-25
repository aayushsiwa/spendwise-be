package handlers

import (
	"encoding/csv"
	"io"
	"net/http"
	"strconv"
	"strings"

	"aayushsiwa/expense-tracker/db"
	"aayushsiwa/expense-tracker/secure"

	"github.com/gin-gonic/gin"
)

func ImportCSV(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV file not provided"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Skip header
	if _, err := reader.Read(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV format"})
		return
	}

	var importedCount int

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading CSV"})
			return
		}

		if len(record) < 5 {
			continue // skip incomplete records
		}

		date := strings.TrimSpace(record[0])
		description := strings.TrimSpace(record[1])
		category := strings.TrimSpace(record[2])
		amountStr := strings.TrimSpace(record[3])
		recordType := strings.TrimSpace(record[4])
		notes := ""
		if len(record) > 5 {
			notes = strings.TrimSpace(record[5])
		}

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			continue
		}

		var categoryID int
		err = db.DB.QueryRow(`SELECT id FROM categories WHERE name = ?`, category).Scan(&categoryID)
		if err != nil {
			res, err := db.DB.Exec(`INSERT INTO categories (name) VALUES (?)`, category)
			if err != nil {
				continue
			}
			lastID, _ := res.LastInsertId()
			categoryID = int(lastID)
		}

		encryptedDescription, _ := secure.Encrypt(description)
		encryptedNotes, _ := secure.Encrypt(notes)

		_, err = db.DB.Exec(`INSERT INTO records (date, description, category_id, amount, type, notes) VALUES (?, ?, ?, ?, ?, ?)`,
			date, encryptedDescription, categoryID, amount, recordType, encryptedNotes)

		if err != nil {
			continue
		}

		importedCount++
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":          "CSV import completed successfully",
		"records_imported": importedCount,
	})
}
