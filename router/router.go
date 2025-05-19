package router

import (
	"database/sql"
	"main/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRouter(db *sql.DB, docRoot string) *gin.Engine {
	// db parameter is kept for backward compatibility but is no longer used
	r := gin.Default()

	// Group all API routes under /api path
	api := r.Group("/api")
	{
		// Document endpoints
		api.POST("/document", handlers.CreateDocument(docRoot))
		api.GET("/document/tree", handlers.GetDocumentTree(docRoot))
		api.POST("/document/update-parent", handlers.UpdateDocumentParent(docRoot))
		
		// Upload endpoint
		api.POST("/upload/:id", handlers.UploadImage(docRoot))
		
		// Library endpoints
		api.POST("/library/create", handlers.CreateLibrary())
		api.GET("/library/list", handlers.ListLibraries(docRoot))
		
		// Library config endpoints
		api.GET("/library/config", handlers.GetLibraryConfig(docRoot))
		api.POST("/library/config", handlers.UpdateLibraryConfig(docRoot))
	}

	return r
}
