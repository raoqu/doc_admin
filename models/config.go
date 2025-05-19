package models

// Config represents library configuration settings
type Config struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Key      string `json:"key"`
	Value    string `json:"value"`
}
