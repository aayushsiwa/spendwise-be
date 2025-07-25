package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"aayushsiwa/expense-tracker/db"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/secure"

	"github.com/gin-gonic/gin"
)

func GetRecords(c *gin.Context) {
	rows, err := db.DB.Query("SELECT * FROM records ORDER BY date DESC")
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

		// Optional: decrypt if storing encrypted data
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

	row := db.DB.QueryRow("SELECT * FROM records WHERE id = ?", id)
	var rec models.Record
	if err := row.Scan(&rec.ID, &rec.Date, &rec.Description, &rec.Category, &rec.Amount, &rec.Type, &rec.Notes); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read record"})
		}
		return
	}

	// Optional: decrypt if storing encrypted data
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

	if rec.Date == "" || rec.Description == "" || rec.Category == "" || rec.Amount <= 0 || rec.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid record fields"})
		return
	}

	// Optional: encrypt if storing encrypted data
	rec.Description, _ = secure.Encrypt(rec.Description)
	rec.Notes, _ = secure.Encrypt(rec.Notes)

	res, err := db.DB.Exec(`
		INSERT INTO records (date, description, category, amount, type, notes)
		VALUES (?, ?, ?, ?, ?, ?)`,
		rec.Date, rec.Description, rec.Category, rec.Amount, rec.Type, rec.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert record"})
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	rec.ID = int(id)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Created with ID - " + strconv.Itoa(rec.ID),
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

	var idC models.Record

	err = db.DB.QueryRow(`
		SELECT id FROM records WHERE id = ?`, id).Scan(&idC.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	_, err = db.DB.Exec(`
		UPDATE records 
		SET date = ?, description = ?, category = ?, amount = ?, type = ?, notes = ?
		WHERE id = ?`,
		rec.Date, rec.Description, rec.Category, rec.Amount, rec.Type, rec.Notes, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update record"})
		return
	}

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

	c.JSON(http.StatusOK, gin.H{
		"message": "Record with id " + strconv.Itoa(id) + " deleted successfully",
	})
}
