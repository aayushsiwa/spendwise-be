package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"aayushsiwa/expense-tracker/db"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/secure"
	"aayushsiwa/expense-tracker/utils"

	"github.com/gin-gonic/gin"
)

func GetRecords(c *gin.Context) {
	rows, err := db.DB.Query(`
		SELECT r.id, r.date, r.description, c.name as category, r.amount, r.type, r.notes
		FROM records r
		JOIN categories c ON r.category_id = c.id
		ORDER BY r.date DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve records"})
		return
	}
	defer rows.Close()

	var records []models.Record
	for rows.Next() {
		var rec models.Record
		err := rows.Scan(&rec.ID, &rec.Date, &rec.Description, &rec.Category, &rec.Amount, &rec.Type, &rec.Notes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read records"})
			return
		}

		rec.Description, _ = secure.Decrypt(rec.Description)
		rec.Notes, _ = secure.Decrypt(rec.Notes)

		records = append(records, rec)
	}

	c.JSON(http.StatusOK, records)
}

func GetRecord(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	row := db.DB.QueryRow(`
		SELECT r.id, r.date, r.description, c.name as category, r.amount, r.type, r.notes
		FROM records r
		JOIN categories c ON r.category_id = c.id
		WHERE r.id = ?
	`, id)

	var rec models.Record
	if err := row.Scan(&rec.ID, &rec.Date, &rec.Description, &rec.Category, &rec.Amount, &rec.Type, &rec.Notes); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read record"})
		}
		return
	}

	rec.Description, _ = secure.Decrypt(rec.Description)
	rec.Notes, _ = secure.Decrypt(rec.Notes)

	c.JSON(http.StatusOK, rec)
}

func CreateRecord(c *gin.Context) {
	var rec models.Record
	if err := c.ShouldBindJSON(&rec); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	if rec.Date == "" || rec.Category == "" || rec.Amount <= 0 || rec.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid record fields"})
		return
	}

	rec.Description, _ = secure.Encrypt(rec.Description)
	rec.Notes, _ = secure.Encrypt(rec.Notes)

	customId, _ := utils.GenerateCustomID(rec.Date)
	rec.ID, _ = strconv.Atoi(customId)

	var categoryId int
	err := db.DB.QueryRow("SELECT id FROM categories WHERE name = ?", rec.Category).Scan(&categoryId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}

	_, execErr := db.DB.Exec(`
		INSERT INTO records (id, date, description, category_id, amount, type, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rec.ID, rec.Date, rec.Description, categoryId, rec.Amount, rec.Type, rec.Notes)
	if execErr != nil {
		log.Println("Error inserting record:", execErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert record"})
		return
	}

	UpdateSummary()

	c.JSON(http.StatusCreated, gin.H{
		"message": "Created with ID - " + customId,
		"id":      rec.ID,
	})
}

func PatchRecord(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var rec models.Record
	if err := c.ShouldBindJSON(&rec); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	if rec.Date == "" || rec.Category == "" || rec.Amount <= 0 || rec.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid fields"})
		return
	}

	rec.Description, _ = secure.Encrypt(rec.Description)
	rec.Notes, _ = secure.Encrypt(rec.Notes)

	var categoryId int
	err = db.DB.QueryRow("SELECT id FROM categories WHERE name = ?", rec.Category).Scan(&categoryId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}

	var exists int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM records WHERE id = ?", id).Scan(&exists)
	if err != nil || exists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	_, err = db.DB.Exec(`
		UPDATE records 
		SET date = ?, description = ?, category_id = ?, amount = ?, type = ?, notes = ?
		WHERE id = ?`,
		rec.Date, rec.Description, categoryId, rec.Amount, rec.Type, rec.Notes, id)
	if err != nil {
		log.Println("Error updating record:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update record"})
		return
	}

	UpdateSummary()

	rec.ID = id
	c.JSON(http.StatusOK, rec)
}

func DeleteRecord(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	res, err := db.DB.Exec(`DELETE FROM records WHERE id = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete record"})
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	UpdateSummary()

	c.JSON(http.StatusOK, gin.H{
		"message": "Record with id " + strconv.Itoa(id) + " deleted successfully",
	})
}
