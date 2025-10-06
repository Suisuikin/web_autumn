package models

import "time"

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"size:150;unique;not null"`
	PasswordHash string `gorm:"size:255;not null"`
	Email        string `gorm:"size:255"`
	IsActive     bool   `gorm:"not null;default:true"`
	IsModerator  bool   `gorm:"not null;default:false"`
}

type Layer struct {
	ID          uint    `gorm:"primaryKey"`
	Name        string  `gorm:"size:255;not null;unique"`
	Description string  `gorm:"type:text"`
	ImageURL    *string `gorm:"size:2048"`
	FromYear    int     `gorm:"not null"`
	ToYear      int     `gorm:"not null"`
	Status      string  `gorm:"size:50;not null;default:'active'"`
}

type ResearchRequest struct {
	ID          uint      `gorm:"primaryKey"`
	Status      string    `gorm:"size:50;not null;default:'draft'"`
	CreatedAt   time.Time `gorm:"not null;autoCreateTime"`
	FormedAt    *time.Time
	CompletedAt *time.Time
	UserID      uint `gorm:"not null"`
	User        User
	Notes       *string
	Layers      []Layer `gorm:"many2many:request_layers;"`
}

type RequestLayer struct {
	ResearchRequestID uint    `gorm:"primaryKey"`
	LayerID           uint    `gorm:"primaryKey"`
	Comment           *string `gorm:"size:1024"`
}
