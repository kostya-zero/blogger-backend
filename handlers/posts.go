package handlers

import "gorm.io/gorm"

type PostsHandler struct {
	DB *gorm.DB
}

func NewPostsHandler(db *gorm.DB) *PostsHandler {
	return &PostsHandler{DB: db}
}
