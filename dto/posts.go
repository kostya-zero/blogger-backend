package dto

type CreatePostRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=128"`
	Description string `json:"description" validate:"required,min=1,max=256"`
	Content     string `json:"content" validate:"required,min=1"`
}
