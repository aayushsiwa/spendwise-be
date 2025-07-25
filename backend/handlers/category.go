package handlers

import (
	"aayushsiwa/expense-tracker/db"
	"aayushsiwa/expense-tracker/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetCategories(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id, name FROM categories ORDER BY name ASC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.Name); err == nil {
			categories = append(categories, cat)
		}
	}
	c.JSON(http.StatusOK, categories)
}

func GetCategoryRecords(c *gin.Context) {
	pathParts := strings.Split(c.Request.URL.Path, "/")
	if len(pathParts) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL path"})
		return
	}
	categoryId := pathParts[len(pathParts)-1]

	var categoryName string
	err := db.DB.QueryRow("SELECT name FROM categories WHERE id = ?", categoryId).Scan(&categoryName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	rows, err := db.DB.Query(`
		SELECT r.id, r.date, r.description, r.category_id, c.name, r.amount, r.type, r.notes
		FROM records r
		JOIN categories c ON r.category_id = c.id
		WHERE c.name = ?
		ORDER BY r.date DESC
	`, categoryName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch records"})
		return
	}
	defer rows.Close()

	var records []models.Record
	for rows.Next() {
		var r models.Record
		err := rows.Scan(&r.ID, &r.Date, &r.Description, &categoryId, &r.Category, &r.Amount, &r.Type, &r.Notes)
		if err == nil {
			records = append(records, r)
		}
	}
	c.JSON(http.StatusOK, records)
}

func CreateCategory(c *gin.Context) {
	var cat models.Category
	if err := c.BindJSON(&cat); err != nil || cat.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	result, err := db.DB.Exec("INSERT INTO categories (name) VALUES (?)", cat.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id, "name": cat.Name})
}

func UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var cat models.Category
	if err := c.BindJSON(&cat); err != nil || cat.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	_, err := db.DB.Exec("UPDATE categories SET name = ? WHERE id = ?", cat.Name, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "name": cat.Name})
}

func DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	_, err := db.DB.Exec("DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
