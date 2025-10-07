package models

type CreateLayerDTO struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description,omitempty"`
	FromYear    int     `json:"from_year" binding:"required"`
	ToYear      int     `json:"to_year" binding:"required"`
	ImageURL    *string `json:"image_url,omitempty"`
	Status      string  `json:"status,omitempty"`
}

type UpdateLayerDTO struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	FromYear    *int    `json:"from_year,omitempty"`
	ToYear      *int    `json:"to_year,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
	Status      *string `json:"status,omitempty"`
}

type LayerCommentDTO struct {
	Comment *string `json:"comment,omitempty"`
}

type CreateRequestDTO struct {
	UserID uint    `json:"user_id" binding:"required"`
	Notes  *string `json:"notes,omitempty"`
}

type UpdateRequestDTO struct {
	Notes  *string `json:"notes,omitempty"`
	Status *string `json:"status,omitempty"`
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
