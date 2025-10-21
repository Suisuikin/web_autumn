package models

type CreateLayerDTO struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description,omitempty"`
	FromYear    int     `json:"from_year" binding:"required"`
	ToYear      int     `json:"to_year" binding:"required"`
	Words       string  `json:"words" binding:"required"` // Список слов через запятую
	ImageURL    *string `json:"image_url,omitempty"`
}

type UpdateLayerDTO struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	FromYear    *int    `json:"from_year,omitempty"`
	ToYear      *int    `json:"to_year,omitempty"`
	Words       *string `json:"words,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
}

type CreateRequestDTO struct {
	TextForAnalysis *string `json:"text_for_analysis,omitempty"`
	Purpose         *string `json:"purpose,omitempty"`
}

type UpdateRequestDTO struct {
	TextForAnalysis *string `json:"text_for_analysis,omitempty"`
	Purpose         *string `json:"purpose,omitempty"`
}

type UpdateLayerCommentDTO struct {
	Comment *string `json:"comment,omitempty"`
}

type RegisterUserDTO struct {
	Username string  `json:"username" binding:"required"`
	Email    *string `json:"email,omitempty"`
	Password string  `json:"password" binding:"required"`
}

type UpdateUserDTO struct {
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
}

type LoginDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CartIconDTO struct {
	RequestID *uint `json:"chrono_id,omitempty"`
	Count     int   `json:"count"`
}
