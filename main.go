package main

import (
	"flag"
	"fmt"
	"main/router"
)

type Config struct {
	DocRoot string `yaml:"doc_root"`
}

func main() {
	// Define command line flags
	dirRootFlag := flag.String("dir", ".", "Document root directory path")
	portFlag := flag.Int("port", 8080, "Port to run the server on")
	
	// Parse command line arguments
	flag.Parse()
	
	// Use the provided values
	dirRoot := *dirRootFlag
	port := *portFlag
	
	fmt.Printf("Starting server with document root: %s on port: %d\n", dirRoot, port)
	
	// Initialize router with the document root path
	r := router.SetupRouter(nil, dirRoot)
	r.Run(fmt.Sprintf(":%d", port))
}
