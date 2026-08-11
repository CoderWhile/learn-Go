package model

import "time"

type Comment struct {
	ID        int
	UserID    int
	UserName  string
	Content   string
	MovieID   int
	ParentID  int
	CreatedAt time.Time
}
type CommentNode struct {
	ID        int            `json:"id"`
	UserName  string         `json:"userName"`
	Content   string         `json:"content"`
	CreatedAt string         `json:"createdAt"`
	Children  []*CommentNode `json:"children"`
}
