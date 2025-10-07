package repository

import (
	"golang.org/x/crypto/bcrypt"
	"rip/internal/app/models"
)

func (r *Repository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *Repository) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *Repository) UpdateUser(id uint, input *models.UpdateUserDTO) error {
	updates := map[string]interface{}{}
	if input.Email != nil {
		updates["email"] = *input.Email
	}
	if input.Username != nil {
		updates["username"] = *input.Username
	}
	if input.Password != nil {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		updates["password_hash"] = string(hashed)
	}
	return r.db.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error
}
