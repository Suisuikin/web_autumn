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
	ID          uint    `gorm:"primaryKey" json:"id"`
	Name        string  `gorm:"size:255;not null;unique" json:"name"`
	Description string  `gorm:"type:text" json:"description,omitempty"`
	ImageURL    *string `gorm:"size:2048" json:"image_url,omitempty"`
	FromYear    int     `gorm:"not null" json:"from_year"`
	ToYear      int     `gorm:"not null" json:"to_year"`
	Status      string  `gorm:"size:50;not null;default:'active'" json:"status"`
}

type ResearchRequest struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Status      string     `gorm:"size:50;not null;default:'draft'" json:"status"`
	CreatedAt   time.Time  `gorm:"not null;autoCreateTime" json:"created_at"`
	FormedAt    *time.Time `json:"formed_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	UserID      uint       `gorm:"not null" json:"user_id"`
	ModeratorID uint       `gorm:"not null" json:" moderator_id"`
	User        User       `json:"user,omitempty"`
	Notes       *string    `json:"notes,omitempty"`
	Layers      []Layer    `gorm:"many2many:request_layers;" json:"layers,omitempty"`
}

type RequestLayer struct {
	ResearchRequestID uint    `gorm:"primaryKey"`
	LayerID           uint    `gorm:"primaryKey"`
	Comment           *string `gorm:"size:1024" json:"comment,omitempty"`
}
