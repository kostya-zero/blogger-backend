package models

import "time"

type Post struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint      `gorm:"not null;index:posts_user_id_idx" json:"-"`
	Title       string    `gorm:"type:text;not null;index:posts_index_0" json:"title"`
	CreatedAt   time.Time `gorm:"type:timestamp;not null;default:now()" json:"created_at"`
	Description string    `gorm:"type:text;not null" json:"description"`
	Content     string    `gorm:"type:text;not null" json:"content"`

	// Relationships
	User  User   `gorm:"foreignKey:UserID" json:"user"`
	Likes []Like `gorm:"foreignKey:PostID" json:"-"`
}
