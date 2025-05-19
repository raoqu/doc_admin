package handlers

import (
	"database/sql"
	"net/http"
	"path/filepath"

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
		
		rows, err := db.Query("SELECT id, title, content, parent_id FROM documents")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var docs []models.Document
		for rows.Next() {
			var doc models.Document
			rows.Scan(&doc.ID, &doc.Title, &doc.Content, &doc.ParentID)
			docs = append(docs, doc)
		}
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
