package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// generateUniqueFilename creates a unique filename with the given extension
func generateUniqueFilename(extension string) string {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%d%s", timestamp, extension)
}

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

		// Determine filename
		filename := file.Filename
		
		// If no filename or it's empty, generate a unique one
		if filename == "" {
			// Try to determine extension from content type
			extension := ".png" // Default extension
			contentType := file.Header.Get("Content-Type")
			if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
				extension = ".jpg"
			} else if strings.Contains(contentType, "png") {
				extension = ".png"
			} else if strings.Contains(contentType, "gif") {
				extension = ".gif"
			}
			
			filename = generateUniqueFilename(extension)
		} else {
			// Ensure filename is safe and unique
			// Remove any path components that might be in the filename
			filename = filepath.Base(filename)
			
			// Check if file already exists, if so, make it unique
			fileExt := filepath.Ext(filename)
			fileBase := strings.TrimSuffix(filename, fileExt)
			
			for i := 1; ; i++ {
				fullPath := filepath.Join(targetDir, filename)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					// File doesn't exist, we can use this name
					break
				}
				// File exists, try a new name
				filename = fmt.Sprintf("%s_%d%s", fileBase, i, fileExt)
			}
		}

		dst := filepath.Join(targetDir, filename)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot save file"})
			return
		}

		// Return relative path for client use
		relativePath := fmt.Sprintf("/pic/%s/%s/%s", libraryName, docID, filename)
		c.JSON(http.StatusOK, gin.H{
			"message": "File uploaded successfully", 
			"path": relativePath,
			"filename": filename,
		})
	}
}

// GetImage serves an image file for a specific document
func GetImage(docRoot string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get parameters from URL
		libraryName := c.Param("library")
		docID := c.Param("docid")
		filename := c.Param("filename")
		
		// Validate parameters
		if libraryName == "" || docID == "" || filename == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Library, document ID, and filename are required"})
			return
		}
		
		// Construct the file path
		libraryPath := filepath.Join(docRoot, libraryName)
		imagePath := filepath.Join(libraryPath, "pic", docID, filename)
		
		// Check if file exists
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
			return
		}
		
		// Serve the file
		c.File(imagePath)
	}
}
