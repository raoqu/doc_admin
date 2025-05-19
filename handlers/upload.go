package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func UploadImage(docRoot string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get library name from query parameter
		libraryName := c.Query("library")
		if libraryName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Library name is required"})
			return
		}
		
		docID := c.Param("id")
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file is received"})
			return
		}

		// Create path within the specific library
		libraryPath := filepath.Join(docRoot, libraryName)
		targetDir := filepath.Join(libraryPath, "pic", docID)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot create directory"})
			return
		}

		dst := filepath.Join(targetDir, file.Filename)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot save file"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "path": dst})
	}
}
