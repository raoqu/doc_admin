package handlers

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

/*
	{
		"name": "mybook1",
		"base_path": "./storage"
	  }
*/
func CreateLibrary() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Name     string `json:"name"`
			BasePath string `json:"base_path"` // 库路径
		}
		var req Req
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Use the base_path directly as the library path
		libPath := req.BasePath
		dbPath := filepath.Join(libPath, "blog.db")
		picPath := filepath.Join(libPath, "pic")

		// Only check if the database file exists, not the directory
		if _, err := os.Stat(dbPath); err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "知识库已存在"})
			return
		}
		
		// If directory exists but no database, that's fine - we'll create the database

		// 创建目录
		if err := os.MkdirAll(picPath, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "目录创建失败"})
			return
		}

		// 创建 SQLite 数据库并初始化表结构
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库创建失败"})
			return
		}
		defer db.Close()

		// Create documents table
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS documents (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				title TEXT,
				content TEXT,
				parent_id INTEGER
			)
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库表初始化失败"})
			return
		}
		
		// Create config table
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS config (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT,
				key TEXT,
				value TEXT
			)
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "配置表初始化失败"})
			return
		}
		
		// Insert blog name into config table
		_, err = db.Exec("INSERT INTO config (name, key, value) VALUES (?, ?, ?)", "blog", "name", req.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "配置初始化失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "知识库创建成功", "name": req.Name, "path": libPath})
	}
}

// ListLibraries returns a list of all library folders in the base path
func ListLibraries(basePath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if base path exists, create it if it doesn't
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			// Create the directory
			if err := os.MkdirAll(basePath, 0755); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create base path"})
				return
			}
			// Return empty list since the directory was just created
			c.JSON(http.StatusOK, gin.H{"libraries": []map[string]string{}})
			return
		}

		// Read only top-level directories in the base path (non-recursive)
		entries, err := os.ReadDir(basePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read directories"})
			return
		}

		// Filter top-level directories that contain a blog.db file (which indicates a library)
		libraries := []map[string]string{}
		// Non-recursive enumeration - only checking direct children of basePath
		for _, entry := range entries {
			if entry.IsDir() {
				libPath := filepath.Join(basePath, entry.Name())
				dbPath := filepath.Join(libPath, "blog.db")
				
				// Check if blog.db exists in this directory
				if _, err := os.Stat(dbPath); err == nil {
					// Open the database to get the blog name from config
					db, err := sql.Open("sqlite3", dbPath)
					if err == nil {
						defer db.Close()
						
						// Try to get blog name from config
						var blogName string
						row := db.QueryRow("SELECT value FROM config WHERE name = 'blog' AND key = 'name' LIMIT 1")
						row.Scan(&blogName)
						
						// If no blog name found, use directory name
						if blogName == "" {
							blogName = entry.Name()
						}
						
						libraries = append(libraries, map[string]string{
							"name": blogName,
							"path": libPath,
							"dir": entry.Name(),
						})
					} else {
						// Fallback if can't open database
						libraries = append(libraries, map[string]string{
							"name": entry.Name(),
							"path": libPath,
							"dir": entry.Name(),
						})
					}
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{"libraries": libraries})
	}
}

// GetLibraryConfig retrieves configuration for a specific library
func GetLibraryConfig(docRoot string) gin.HandlerFunc {
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
		
		// Query all config entries
		rows, err := db.Query("SELECT id, name, key, value FROM config")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query config"})
			return
		}
		defer rows.Close()
		
		// Build config map
		config := make(map[string]map[string]string)
		for rows.Next() {
			var id int64
			var name, key, value string
			if err := rows.Scan(&id, &name, &key, &value); err != nil {
				continue
			}
			
			// Initialize the map for this name if it doesn't exist
			if _, ok := config[name]; !ok {
				config[name] = make(map[string]string)
			}
			
			// Add the key-value pair
			config[name][key] = value
		}
		
		c.JSON(http.StatusOK, gin.H{"config": config})
	}
}

// UpdateLibraryConfig updates configuration for a specific library
func UpdateLibraryConfig(docRoot string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get library name from query parameter
		libraryName := c.Query("library")
		if libraryName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Library name is required"})
			return
		}
		
		// Parse request body
		type ConfigRequest struct {
			Name  string `json:"name"`
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		
		var req ConfigRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Validate request
		if req.Name == "" || req.Key == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name and key are required"})
			return
		}
		
		// Open connection to the library's database
		db, err := getLibraryDB(docRoot, libraryName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to library database"})
			return
		}
		defer db.Close()
		
		// Check if config exists
		var count int
		row := db.QueryRow("SELECT COUNT(*) FROM config WHERE name = ? AND key = ?", req.Name, req.Key)
		if err := row.Scan(&count); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query config"})
			return
		}
		
		var result sql.Result
		if count > 0 {
			// Update existing config
			result, err = db.Exec("UPDATE config SET value = ? WHERE name = ? AND key = ?", req.Value, req.Name, req.Key)
		} else {
			// Insert new config
			result, err = db.Exec("INSERT INTO config (name, key, value) VALUES (?, ?, ?)", req.Name, req.Key, req.Value)
		}
		
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update config"})
			return
		}
		
		rowsAffected, _ := result.RowsAffected()
		c.JSON(http.StatusOK, gin.H{
			"message": "Config updated successfully",
			"updated": rowsAffected > 0,
		})
	}
}
