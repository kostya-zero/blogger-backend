package models

import (
	"time"
)

type User struct {
	ID           uint      `gorm:"primaryKey;autoInrement" json:"-"`
	Name         *string   `gorm:"type:text"`
	Username     string    `gorm:"type:text;unique;not null;index:users_index_0"`
	Email        string    `gorm:"type:text;unique;not null" json:"-"`
	About        *string   `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"type:timestamp;not null;default:now()"`
	PasswordHash string    `gorm:"type:text;not null" json:"-"`
	RefreshToken *string   `gorm:"type:text" json:"-"`

	// Relationships
	Posts []Post `gorm:"foreignKey:UserID" json:"-"`
	Likes []Like `gorm:"foreignKey:UserID" json:"-"`
}
