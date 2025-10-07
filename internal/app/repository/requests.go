package repository

import (
	"errors"
	"rip/internal/app/models"
	"time"

	"gorm.io/gorm"
)

type RequestsRepository struct {
	db *gorm.DB
}

func NewRequestsRepository(db *gorm.DB) *RequestsRepository {
	return &RequestsRepository{db: db}
}

func (r *RequestsRepository) GetOpenRequest(userID uint) (*models.ResearchRequest, error) {
	var req models.ResearchRequest
	err := r.db.Preload("Layers").Where("user_id = ? AND status = ?", userID, "draft").First(&req).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &req, err
}

func (r *RequestsRepository) ListRequests() ([]models.ResearchRequest, error) {
	var requests []models.ResearchRequest
	if err := r.db.Preload("Layers").Find(&requests).Error; err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *RequestsRepository) AddStrategyToDraft(userID, layerID uint) (*models.ResearchRequest, error) {
	req, err := r.GetOpenRequest(userID)
	if err != nil {
		return nil, err
	}

	if req == nil {
		req = &models.ResearchRequest{UserID: userID, Status: "draft"}
		if err := r.db.Create(req).Error; err != nil {
			return nil, err
		}
	}

	layer := models.Layer{}
	if err := r.db.First(&layer, layerID).Error; err != nil {
		return nil, err
	}

	if err := r.db.Model(req).Association("Layers").Append(&layer); err != nil {
		return nil, err
	}

	return req, nil
}

func (r *RequestsRepository) GetRequestByID(id uint) (*models.ResearchRequest, error) {
	var req models.ResearchRequest
	err := r.db.Preload("Layers").First(&req, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &req, err
}

func (r *RequestsRepository) UpdateRequest(id uint, input *models.ResearchRequest) error {
	return r.db.Model(&models.ResearchRequest{}).Where("id = ?", id).Updates(input).Error
}

func (r *RequestsRepository) FormRequest(id uint) error {
	return r.db.Model(&models.ResearchRequest{}).Where("id = ?", id).Update("status", "formed").Error
}

func (r *RequestsRepository) ResolveRequest(id uint) error {
	return r.db.Model(&models.ResearchRequest{}).Where("id = ?", id).Update("status", "resolved").Error
}

func (r *RequestsRepository) UpdateRequestStrategy(reqID, layerID uint, comment string) error {
	return r.db.Model(&models.RequestLayer{}).
		Where("research_request_id = ? AND layer_id = ?", reqID, layerID).
		Update("comment", comment).Error
}

func (r *RequestsRepository) RemoveRequestStrategy(reqID, layerID uint) error {
	return r.db.Where("research_request_id = ? AND layer_id = ?", reqID, layerID).
		Delete(&models.RequestLayer{}).Error
}

func (r *Repository) ListRequests(status, fromStr, toStr string) ([]models.ResearchRequest, error) {
	var requests []models.ResearchRequest
	query := r.db.Preload("Layers")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if fromStr != "" {
		if fromTime, err := time.Parse("2006-01-02", fromStr); err == nil {
			query = query.Where("created_at >= ?", fromTime)
		}
	}

	if toStr != "" {
		if toTime, err := time.Parse("2006-01-02", toStr); err == nil {
			query = query.Where("created_at <= ?", toTime)
		}
	}

	err := query.Find(&requests).Error
	return requests, err
}

func (r *Repository) GetCart() (*models.ResearchRequest, error) {
	var req models.ResearchRequest
	err := r.db.Preload("Layers").Where("status = ?", "черновик").First(&req).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			req = models.ResearchRequest{
				Status: "черновик",
				UserID: 1,
			}
			if err := r.db.Create(&req).Error; err != nil {
				return nil, err
			}
			if err := r.db.Preload("Layers").First(&req, req.ID).Error; err != nil {
				return nil, err
			}
			return &req, nil
		}
		return nil, err
	}
	return &req, nil
}

func (r *Repository) AddLayerToDraft(layerID uint) (*models.ResearchRequest, error) {
	req := models.ResearchRequest{Status: "draft"}
	layer := models.Layer{ID: layerID}
	req.Layers = append(req.Layers, layer)
	if err := r.db.Create(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *Repository) AddLayerToOpenRequest(layerID uint) (*models.ResearchRequest, error) {
	var req models.ResearchRequest
	err := r.db.Preload("Layers").Where("status = ?", "черновик").First(&req).Error
	if err != nil {
		req = models.ResearchRequest{Status: "открыта"}
		if err := r.db.Create(&req).Error; err != nil {
			return nil, err
		}
	}

	layer := models.Layer{ID: layerID}
	if err := r.db.Model(&req).Association("Layers").Append(&layer); err != nil {
		return nil, err
	}

	if err := r.db.Preload("Layers").First(&req, req.ID).Error; err != nil {
		return nil, err
	}

	return &req, nil
}

func (r *Repository) GetRequest(id uint) (*models.ResearchRequest, error) {
	var req models.ResearchRequest
	err := r.db.Preload("Layers").First(&req, id).Error
	return &req, err
}

func (r *Repository) UpdateRequest(id uint, input *models.ResearchRequest) error {
	return r.db.Model(&models.ResearchRequest{}).Where("id = ?", id).Updates(input).Error
}

func (r *Repository) FormRequest(id uint) error {
	var req models.ResearchRequest
	err := r.db.First(&req, id).Error
	if err != nil {
		return err
	}

	if req.Status != "черновик" {
		return errors.New("можно формировать только заявки в статусе черновик")
	}

	return r.db.Model(&req).Update("status", "рассматривается").Error
}

func (r *Repository) ResolveRequest(id uint) error {
	var req models.ResearchRequest
	err := r.db.First(&req, id).Error
	if err != nil {
		return err
	}

	if req.Status != "рассматривается" {
		return errors.New("можно рассматривать только заявки в статусе рассматривается")
	}

	return r.db.Model(&req).Update("status", "рассмотрен").Error
}

func (r *Repository) DeleteRequest(id uint) error {
	var req models.ResearchRequest
	err := r.db.First(&req, id).Error
	if err != nil {
		return err
	}

	if req.Status != "черновик" {
		return errors.New("можно удалять только заявки в статусе черновик")
	}

	return r.db.Delete(&models.ResearchRequest{}, id).Error
}

func (r *Repository) UpdateRequestLayer(reqID, layerID uint, comment *string) error {
	var reqLayer models.RequestLayer

	err := r.db.Where("research_request_id = ? AND layer_id = ?", reqID, layerID).First(&reqLayer).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			reqLayer = models.RequestLayer{
				ResearchRequestID: reqID,
				LayerID:           layerID,
				Comment:           comment,
			}
			return r.db.Create(&reqLayer).Error
		}
		return err
	}

	reqLayer.Comment = comment
	return r.db.Save(&reqLayer).Error
}

func (r *Repository) RemoveLayerFromRequest(requestID, layerID uint) error {
	return r.db.Model(&models.ResearchRequest{ID: requestID}).Association("Layers").Delete(&models.Layer{ID: layerID})
}
