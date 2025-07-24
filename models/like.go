package models

type Like struct {
    PostID uint `gorm:"not null;primaryKey"`
    UserID uint `gorm:"not null;primaryKey;index:likes_user_id_idx"`
    
    // Relationships
    Post Post `gorm:"foreignKey:PostID"`
    User User `gorm:"foreignKey:UserID"`
}
