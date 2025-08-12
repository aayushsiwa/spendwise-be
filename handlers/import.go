package handlers

import (
	"encoding/csv"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"aayushsiwa/expense-tracker/db"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/secure"
	"aayushsiwa/expense-tracker/utils"

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
			continue
		}

		dateStr := strings.TrimSpace(record[0])
		description := strings.TrimSpace(record[1])
		category := strings.ToLower(strings.TrimSpace(record[2]))
		amountStr := strings.TrimSpace(record[3])
		recordType := strings.TrimSpace(record[4])
		note := ""
		if len(record) > 5 {
			note = strings.TrimSpace(record[5])
		}

		// Validate date
		date, err := utils.ParseDate(dateStr)
		if err != nil {
			continue
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
		encryptedNote, _ := secure.Encrypt(note)

		_, err = db.DB.Exec(`INSERT INTO records (date, description, category_id, amount, type, note) VALUES (?, ?, ?, ?, ?, ?)`,
			date, encryptedDescription, categoryID, amount, recordType, encryptedNote)
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

func ImportJSON(c *gin.Context) {
	var records []models.Record
	if err := c.ShouldBindJSON(&records); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON array"})
		return
	}

	importedCount := 0

	for _, rec := range records {
		if rec.Date == "" || rec.Description == "" || rec.Category == "" || rec.Type == "" {
			continue
		}

		// Normalize category
		category := strings.ToLower(strings.TrimSpace(rec.Category))
		dateStr := strings.TrimSpace(rec.Date)

		date, err := utils.ParseDate(dateStr)
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

		encDesc, _ := secure.Encrypt(rec.Description)
		encNote, _ := secure.Encrypt(rec.Note)

		customID, err := utils.GenerateCustomID(date)
		if err != nil {
			continue
		}

		_, err = db.DB.Exec(`INSERT INTO records (id, date, description, category_id, amount, type, note) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			customID, date, encDesc, categoryID, rec.Amount, rec.Type, encNote)
		if err != nil {
			log.Printf("Failed to insert record: %v", err)
			continue
		}

		importedCount++
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":          "JSON import completed successfully",
		"records_imported": importedCount,
	})
}
