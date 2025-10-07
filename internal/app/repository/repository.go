package repository

import (
	"errors"
	"fmt"
	"log"
	"rip/internal/app/models"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetLayers() ([]models.Layer, error) {
	var layers []models.Layer
	if err := r.db.Find(&layers).Error; err != nil {
		return nil, fmt.Errorf("не удалось получить слои: %w", err)
	}
	if len(layers) == 0 {
		return nil, fmt.Errorf("нет данных в таблице layers")
	}

	if err := r.db.Find(&layers).Error; err != nil {
		return nil, fmt.Errorf("не удалось получить слои: %w", err)
	}
	log.Println("Found layers:", len(layers))

	return layers, nil
}

func (r *Repository) GetLayerByID(id uint) (*models.Layer, error) {
	var layer models.Layer
	if err := r.db.First(&layer, id).Error; err != nil {
		return nil, fmt.Errorf("слой с ID %d не найден: %w", id, err)
	}
	return &layer, nil
}

func (r *Repository) SearchLayers(query string) ([]models.Layer, error) {
	var layers []models.Layer
	if query == "" {
		return r.GetLayers()
	}

	if err := r.db.Where("LOWER(name) LIKE ?", "%"+query+"%").Find(&layers).Error; err != nil {
		return nil, fmt.Errorf("ошибка поиска: %w", err)
	}

	return layers, nil
}

func (r *Repository) GetOpenRequest(userID uint) (*models.ResearchRequest, error) {
	var req models.ResearchRequest

	err := r.db.
		Preload("Layers").
		Where("user_id = ? AND status = ?", userID, "черновик").
		Order("id DESC").
		First(&req).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка при получении заявки: %w", err)
	}

	return &req, nil
}

func (r *Repository) CreateNewRequest(userID uint) (*models.ResearchRequest, error) {
	req := models.ResearchRequest{
		Status:    "черновик",
		CreatedAt: time.Now(),
		UserID:    userID,
	}
	if err := r.db.Create(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *Repository) AddLayerToRequest(reqID, layerID uint) error {
	return r.db.Exec("INSERT INTO request_layers (research_request_id, layer_id) VALUES (?, ?) ON CONFLICT DO NOTHING", reqID, layerID).Error
}

func (r *Repository) DeleteLayerFromRequest(reqID, layerID uint) error {
	return r.db.Exec("DELETE FROM request_layers WHERE research_request_id = ? AND layer_id = ?", reqID, layerID).Error
}

func (r *Repository) GetRequestLayers(reqID uint) ([]models.Layer, error) {
	var req models.ResearchRequest
	if err := r.db.Preload("Layers").First(&req, reqID).Error; err != nil {
		return nil, err
	}
	return req.Layers, nil
}

func (r *Repository) FindOrCreateOpenRequest(userID uint) (*models.ResearchRequest, error) {
	var req models.ResearchRequest

	err := r.db.
		Where("user_id = ? AND status = ?", userID, "черновик").
		Order("created_at DESC").
		Preload("Layers").
		First(&req).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			req = models.ResearchRequest{
				UserID: userID,
				Status: "черновик",
			}
			if err := r.db.Create(&req).Error; err != nil {
				return nil, fmt.Errorf("не удалось создать заявку: %w", err)
			}
			return &req, nil
		}
		return nil, fmt.Errorf("ошибка поиска заявки: %w", err)
	}

	return &req, nil
}

func (r *Repository) CloseRequest(requestID uint) error {
	return r.db.Model(&models.ResearchRequest{}).
		Where("id = ?", requestID).
		Update("status", "удалена").Error
}

func (r *Repository) UpdateLayerComments(requestID uint, comments map[uint]string) error {
	for layerID, comment := range comments {
		if err := r.db.Model(&models.RequestLayer{}).
			Where("research_request_id = ? AND layer_id = ?", requestID, layerID).
			Update("comment", comment).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) UpdateRequestStatus(requestID uint, status string) error {
	return r.db.Model(&models.ResearchRequest{}).Where("id = ?", requestID).Update("status", status).Error
}

func (r *Repository) GetRequestByID(requestID uint) (*models.ResearchRequest, error) {
	var req models.ResearchRequest
	err := r.db.Preload("Layers").First(&req, requestID).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *Repository) GetRequestWithComments(requestID uint) (*models.ResearchRequest, error) {
	var req models.ResearchRequest
	err := r.db.Preload("Layers").
		Preload("Layers.RequestLayers", "research_request_id = ?", requestID).
		First(&req, requestID).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *Repository) GetLayerComments(requestID uint) (map[uint]string, error) {
	var rows []models.RequestLayer
	err := r.db.Where("research_request_id = ?", requestID).Find(&rows).Error
	if err != nil {
		return nil, err
	}

	comments := make(map[uint]string)
	for _, rl := range rows {
		if rl.Comment != nil {
			comments[rl.LayerID] = *rl.Comment
		}
	}
	return comments, nil
}

func (r *Repository) UpdateRequestNotes(requestID uint, notes string) error {
	return r.db.Model(&models.ResearchRequest{}).
		Where("id = ?", requestID).
		Update("notes", notes).Error
}
