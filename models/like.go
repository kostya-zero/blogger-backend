package models

type Like struct {
	PostID uint `gorm:"not null;primaryKey" json:"-"`
	UserID uint `gorm:"not null;primaryKey;index:likes_user_id_idx" json:"-"`

	// Relationships
	Post Post `gorm:"foreignKey:PostID;references:ID"`
	User User `gorm:"foreignKey:UserID;references:ID"`
}
