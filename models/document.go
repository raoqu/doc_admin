package models

type Document struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	ParentID int64  `json:"parent_id"` // 用于树状结构
}
