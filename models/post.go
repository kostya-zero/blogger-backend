package models

import "time"

type Post struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	UserID      uint      `gorm:"not null;index:posts_user_id_idx"`
	Title       string    `gorm:"type:text;not null;index:posts_index_0"`
	CreatedAt   time.Time `gorm:"type:timestamp;not null;default:now()"`
	Description string    `gorm:"type:text;not null"`
	Content     string    `gorm:"type:text;not null"`

	// Relationships
	User  User   `gorm:"foreignKey:UserID"`
	Likes []Like `gorm:"foreignKey:PostID"`
}
