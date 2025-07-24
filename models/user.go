package models

import (
	"time"
)

type User struct {
	ID           uint      `gorm:"primaryKey;autoInrement"`
	Name         string    `gorm:"type:text;not null;index:users_index_0"`
	Email        string    `gorm:"type:text;unique;not null"`
	About        *string   `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"type:timestamp;not null;default:now()"`
	PasswordHash string    `gorm:"type:text;not null"`
	RefreshToken *string   `gorm:"type:text"`

	// Relationships
	Posts []Post `gorm:"foreignKey:UserID"`
	Likes []Like `gorm:"foreignKey:UserID"`
}
