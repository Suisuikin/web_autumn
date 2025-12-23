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
    // Если хотите, чтобы описание всегда выводилось, уберите omitempty
    Description string  `gorm:"type:text" json:"description"`
    ImageURL    *string `gorm:"size:2048" json:"image_url"` // Убрал omitempty
    FromYear    int     `gorm:"not null" json:"from_year"`
    ToYear      int     `gorm:"not null" json:"to_year"`
    Words       string  `gorm:"type:text;not null" json:"words"`
    Status      string  `gorm:"size:50;not null;default:'active'" json:"status"`
}

type ResearchRequest struct {
    ID          uint       `gorm:"primaryKey" json:"id"`
    Status      string     `gorm:"size:50;not null;default:'draft'" json:"status"`
    CreatedAt   time.Time  `gorm:"not null;autoCreateTime" json:"created_at"`
    FormedAt    *time.Time `json:"formed_at"` // Убрал omitempty (будет null, если не сформировано)
    CompletedAt *time.Time `json:"completed_at"` // Убрал omitempty
    UserID      uint       `gorm:"not null" json:"user_id"`
    ModeratorID *uint      `json:"moderator_id"` // Убрал omitempty

    // ВАЖНО: Убрал omitempty, теперь поле будет в JSON всегда (null или строка)
    TextForAnalysis *string `gorm:"type:text" json:"text_for_analysis"`
    Purpose         *string `gorm:"type:text" json:"purpose"`

    ResultFromYear *int `json:"result_from_year"` // Убрал omitempty
    ResultToYear   *int `json:"result_to_year"`   // Убрал omitempty
    MatchedLayers  *int `json:"matched_layers"`   // Убрал omitempty

    // Layers всегда будет в ответе (null или массив)
    Layers []Layer `gorm:"many2many:request_layers;" json:"layers"`
}

type RequestLayer struct {
    ResearchRequestID uint    `gorm:"primaryKey"`
    LayerID           uint    `gorm:"primaryKey"`
    MatchCount        int     `gorm:"default:0" json:"match_count"`
    Comment           *string `gorm:"size:1024" json:"comment"` // Убрал omitempty
}

type AsyncChronoRequest struct {
    ResearchRequestID uint   `json:"research_request_id" binding:"required"`
    AuthToken         string `json:"auth_token" binding:"required"`
}

type AsyncChronoData struct {
    ResearchRequestID uint    `json:"research_request_id"`
    TextForAnalysis   *string `json:"text_for_analysis"`
    Purpose           *string `json:"purpose"`
    UserID            uint    `json:"user_id"`
}

type AsyncResultDTO struct {
    ResearchRequestID uint   `json:"research_request_id" binding:"required"`
    ResultFromYear    *int   `json:"result_from_year"`
    ResultToYear      *int   `json:"result_to_year"`
    MatchedLayers     *int   `json:"matched_layers"`
    AuthToken         string `json:"auth_token" binding:"required"`
}

func (RequestLayer) TableName() string {
    return "request_layers"
}
