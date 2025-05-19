package main

import (
	"log"
	"os"

	"io/ioutil"

	"gopkg.in/yaml.v2"

	"main/router"
)

type Config struct {
	DocRoot string `yaml:"doc_root"`
}

func loadConfig() Config {
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatal("Cannot read config.yaml")
	}
	var cfg Config
	yaml.Unmarshal(data, &cfg)
	return cfg
}

func main() {
	cfg := loadConfig()

	// Ensure the storage directory exists
	if err := os.MkdirAll(cfg.DocRoot, 0755); err != nil {
		log.Fatal("Failed to create storage directory:", err)
	}

	// Initialize router with just the document root path
	r := router.SetupRouter(nil, cfg.DocRoot)
	r.Run(":8080")
}
