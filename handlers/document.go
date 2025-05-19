package handlers

import (
	"database/sql"
	"net/http"
	"path/filepath"
	"strings"

	"main/models"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// getLibraryDB opens a connection to the specified library's database
func getLibraryDB(docRoot string, libraryName string) (*sql.DB, error) {
	libPath := filepath.Join(docRoot, libraryName)
	dbPath := filepath.Join(libPath, "blog.db")
	return sql.Open("sqlite3", dbPath)
}

func CreateDocument(docRoot string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get library name from query parameter
		libraryName := c.Query("library")
		if libraryName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Library name is required"})
			return
		}
		
		// Open connection to the library's database
		db, err := getLibraryDB(docRoot, libraryName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to library database"})
			return
		}
		defer db.Close()
		
		var doc models.Document
		if err := c.ShouldBindJSON(&doc); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res, err := db.Exec("INSERT INTO documents (title, content, parent_id) VALUES (?, ?, ?)", doc.Title, doc.Content, doc.ParentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		id, _ := res.LastInsertId()
		c.JSON(http.StatusOK, gin.H{"id": id})
	}
}

func GetDocumentTree(docRoot string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get library name from query parameter
		libraryName := c.Query("library")
		if libraryName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Library name is required"})
			return
		}
		
		// Open connection to the library's database
		db, err := getLibraryDB(docRoot, libraryName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to library database"})
			return
		}
		defer db.Close()
		
		// Initialize empty docs array
		docs := []models.Document{}
		
		// Check if the documents table exists
		var count int
		row := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='documents'")
		if err := row.Scan(&count); err != nil {
			// For any error, return an error response
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
			return
		}
		
		// If table doesn't exist, return empty array
		if count == 0 {
			c.JSON(http.StatusOK, docs)
			return
		}
		
		// Table exists, query the documents
		rows, err := db.Query("SELECT id, title, content, parent_id FROM documents")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var doc models.Document
			if err := rows.Scan(&doc.ID, &doc.Title, &doc.Content, &doc.ParentID); err != nil {
				continue // Skip documents with scan errors
			}
			docs = append(docs, doc)
		}
		
		// Always return a JSON array, even if empty
		c.JSON(http.StatusOK, docs) // 可以递归构造树结构
	}
}

func UpdateDocumentParent(docRoot string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get library name from query parameter
		libraryName := c.Query("library")
		if libraryName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Library name is required"})
			return
		}
		
		// Open connection to the library's database
		db, err := getLibraryDB(docRoot, libraryName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to library database"})
			return
		}
		defer db.Close()
		
		type UpdateRequest struct {
			ID       int64 `json:"id"`
			ParentID int64 `json:"parent_id"`
		}

		var req UpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 防止将节点拖动到自己或其子节点下（避免递归死循环结构）
		if req.ID == req.ParentID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "不能将节点移动到自身下"})
			return
		}

		_, err = db.Exec("UPDATE documents SET parent_id = ? WHERE id = ?", req.ParentID, req.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "父目录更新成功"})
	}
}

// GetDocumentByID retrieves a document by its ID
func GetDocumentByID(docRoot string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get library name from query parameter
		libraryName := c.Query("library")
		if libraryName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Library name is required"})
			return
		}
		
		// Get document ID from query parameter
		docID := c.Query("id")
		if docID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
			return
		}
		
		// Open connection to the library's database
		db, err := getLibraryDB(docRoot, libraryName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to library database"})
			return
		}
		defer db.Close()
		
		// Check if the documents table exists
		var count int
		row := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='documents'")
		if err := row.Scan(&count); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
			return
		}
		
		// If table doesn't exist, return not found
		if count == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
			return
		}
		
		// Query for the document
		var doc models.Document
		row = db.QueryRow("SELECT id, title, content, parent_id FROM documents WHERE id = ?", docID)
		if err := row.Scan(&doc.ID, &doc.Title, &doc.Content, &doc.ParentID); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
			}
			return
		}
		
		// Return the document
		c.JSON(http.StatusOK, doc)
	}
}

// UpdateDocument updates a document's title and/or content
func UpdateDocument(docRoot string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get library name from query parameter
		libraryName := c.Query("library")
		if libraryName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Library name is required"})
			return
		}
		
		// Open connection to the library's database
		db, err := getLibraryDB(docRoot, libraryName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to library database"})
			return
		}
		defer db.Close()
		
		// Define request structure
		type UpdateRequest struct {
			ID      int64  `json:"id"`
			Title   string `json:"title"`   // Optional
			Content string `json:"content"` // Optional
		}
		
		var req UpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Validate request
		if req.ID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
			return
		}
		
		// Check if document exists
		var exists bool
		row := db.QueryRow("SELECT EXISTS(SELECT 1 FROM documents WHERE id = ?)", req.ID)
		if err := row.Scan(&exists); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check document existence"})
			return
		}
		
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
			return
		}
		
		// Build update query based on provided fields
		var updateFields []string
		var args []interface{}
		
		if req.Title != "" {
			updateFields = append(updateFields, "title = ?")
			args = append(args, req.Title)
		}
		
		if req.Content != "" {
			updateFields = append(updateFields, "content = ?")
			args = append(args, req.Content)
		}
		
		// If no fields to update, return early
		if len(updateFields) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
			return
		}
		
		// Build and execute the query
		query := "UPDATE documents SET " + strings.Join(updateFields, ", ") + " WHERE id = ?"
		args = append(args, req.ID)
		
		result, err := db.Exec(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update document"})
			return
		}
		
		rowsAffected, _ := result.RowsAffected()
		c.JSON(http.StatusOK, gin.H{
			"message": "Document updated successfully",
			"updated": rowsAffected > 0,
		})
	}
}
