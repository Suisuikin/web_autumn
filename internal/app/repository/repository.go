package repository

import (
	"context"
	"errors"
	"log"
	"rip/internal/app/models"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/gorm"
)

type Repository struct {
	db     *gorm.DB
	Minio  *minio.Client
	Bucket string
}

func (r *Repository) GetLayers(query string) ([]models.Layer, error) {
	var layers []models.Layer
	db := r.db.Where("status != ?", "deleted")

	if query != "" {
		db = db.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(query)+"%")
	}

	if err := db.Find(&layers).Error; err != nil {
		return nil, err
	}
	return layers, nil
}

func (r *Repository) GetLayerByID(id uint) (*models.Layer, error) {
	var layer models.Layer
	err := r.db.Where("status != ?", "deleted").First(&layer, id).Error
	if err != nil {
		return nil, err
	}
	return &layer, nil
}

func (r *Repository) CreateLayer(layer *models.Layer) error {
	layer.Status = "active"
	return r.db.Create(layer).Error
}

func (r *Repository) UpdateLayer(id uint, layer *models.Layer) error {
	return r.db.Model(&models.Layer{}).Where("id = ?", id).Updates(layer).Error
}

func (r *Repository) DeleteLayer(id uint) error {
	return r.db.Model(&models.Layer{}).Where("id = ?", id).Update("status", "deleted").Error
}

func (r *Repository) GetCartIcon(userID uint) (*models.CartIconDTO, error) {
	var req models.ResearchRequest
	err := r.db.
		Where("user_id = ? AND status = ?", userID, "draft").
		First(&req).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.CartIconDTO{RequestID: nil, Count: 0}, nil
		}
		return nil, err
	}

	var count int64
	r.db.Model(&models.RequestLayer{}).Where("research_request_id = ?", req.ID).Count(&count)

	return &models.CartIconDTO{
		RequestID: &req.ID,
		Count:     int(count),
	}, nil
}

func (r *Repository) GetRequests(userID uint, isModerator bool, status, dateFrom, dateTo string) ([]models.ResearchRequest, error) {
	db := r.db.Where("status NOT IN (?)", []string{"draft", "deleted"})

	if !isModerator {
		db = db.Where("user_id = ?", userID)
	}

	if status != "" {
		db = db.Where("status = ?", status)
	}

	if dateFrom != "" && dateTo != "" {
		db = db.Where("formed_at BETWEEN ? AND ?", dateFrom, dateTo)
	}

	var requests []models.ResearchRequest
	err := db.Find(&requests).Error

	for i := range requests {
		var matchedCount int64
		r.db.Model(&models.RequestLayer{}).
			Where("research_request_id = ?", requests[i].ID).
			Count(&matchedCount)
		count := int(matchedCount)
		requests[i].MatchedLayers = &count
	}

	return requests, err
}

func (r *Repository) GetRequestByID(id uint, userID uint, isModerator bool) (*models.ResearchRequest, error) {
	var req models.ResearchRequest
	db := r.db

	if !isModerator {
		db = db.Where("user_id = ?", userID)
	}

	db = db.Where("status != ?", "deleted")

	err := db.Preload("Layers").First(&req, id).Error
	if err != nil {
		return nil, err
	}

	return &req, nil
}

func (r *Repository) CreateDraftRequest(userID uint) (*models.ResearchRequest, error) {
	req := models.ResearchRequest{
		Status:    "draft",
		CreatedAt: time.Now(),
		UserID:    userID,
	}
	err := r.db.Create(&req).Error
	return &req, err
}

func (r *Repository) UpdateRequest(id uint, dto *models.UpdateRequestDTO) error {
	updates := make(map[string]interface{})

	if dto.TextForAnalysis != nil {
		updates["text_for_analysis"] = *dto.TextForAnalysis
	}
	if dto.Purpose != nil {
		updates["purpose"] = *dto.Purpose
	}

	return r.db.Model(&models.ResearchRequest{}).Where("id = ? AND status = ?", id, "draft").Updates(updates).Error
}

func (r *Repository) FormRequest(id uint, userID uint) error {
	now := time.Now()
	return r.db.Model(&models.ResearchRequest{}).
		Where("id = ? AND user_id = ? AND status = ?", id, userID, "draft").
		Updates(map[string]interface{}{
			"status":    "formed",
			"formed_at": now,
		}).Error
}

func (r *Repository) CompleteRequest(id uint, moderatorID uint) error {
	var req models.ResearchRequest
	if err := r.db.First(&req, id).Error; err != nil {
		return err
	}

	if req.Status != "formed" {
		return errors.New("заявка должна быть в статусе 'formed'")
	}

	if req.TextForAnalysis == nil {
		return errors.New("текст для анализа отсутствует")
	}

	text := strings.ToLower(*req.TextForAnalysis)
	textWords := strings.Fields(text)

	var layers []models.Layer
	r.db.Where("status = ?", "active").Find(&layers)

	minYear := 3000
	maxYear := 0
	matchedLayersCount := 0

	for _, layer := range layers {
		layerWords := strings.FieldsFunc(layer.Words, func(r rune) bool {
			return r == ',' || r == ' ' || r == ';'
		})

		matchCount := 0
		for _, lw := range layerWords {
			lw = strings.TrimSpace(strings.ToLower(lw))
			if lw == "" {
				continue
			}

			for _, tw := range textWords {
				if tw == lw {
					matchCount++
					break
				}
			}
		}

		if matchCount > 0 {
			r.db.Where("research_request_id = ? AND layer_id = ?", id, layer.ID).
				Assign(models.RequestLayer{
					ResearchRequestID: id,
					LayerID:           layer.ID,
					MatchCount:        matchCount,
				}).
				FirstOrCreate(&models.RequestLayer{})

			matchedLayersCount++

			if layer.FromYear < minYear {
				minYear = layer.FromYear
			}
			if layer.ToYear > maxYear {
				maxYear = layer.ToYear
			}
		}
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":         "completed",
		"completed_at":   now,
		"moderator_id":   moderatorID,
		"matched_layers": matchedLayersCount,
	}
	if matchedLayersCount > 0 {
		updates["result_from_year"] = minYear
		updates["result_to_year"] = maxYear
	}

	return r.db.Model(&models.ResearchRequest{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) DeleteRequest(id uint, userID uint) error {
	query := `UPDATE research_requests SET status = 'deleted' WHERE id = ? AND user_id = ? AND status = 'draft'`
	result := r.db.Exec(query, id, userID)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("заявка не найдена или не может быть удалена")
	}
	return nil
}

func (r *Repository) AddLayerToRequest(userID, layerID uint) error {
	var req models.ResearchRequest
	err := r.db.Where("user_id = ? AND status = ?", userID, "draft").First(&req).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		req = models.ResearchRequest{
			Status:    "draft",
			CreatedAt: time.Now(),
			UserID:    userID,
		}
		if err := r.db.Create(&req).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	requestLayer := models.RequestLayer{
		ResearchRequestID: req.ID,
		LayerID:           layerID,
		MatchCount:        0,
	}

	return r.db.FirstOrCreate(&requestLayer, "research_request_id = ? AND layer_id = ?", req.ID, layerID).Error
}

func (r *Repository) RemoveLayerFromRequest(requestID, layerID uint) error {
	return r.db.Where("research_request_id = ? AND layer_id = ?", requestID, layerID).
		Delete(&models.RequestLayer{}).Error
}

func (r *Repository) UpdateLayerComment(requestID, layerID uint, dto *models.UpdateLayerCommentDTO) error {
	updates := map[string]interface{}{}
	if dto.Comment != nil {
		updates["comment"] = *dto.Comment
	}

	return r.db.Model(&models.RequestLayer{}).
		Where("research_request_id = ? AND layer_id = ?", requestID, layerID).
		Updates(updates).Error
}

func (r *Repository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *Repository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *Repository) UpdateUser(id uint, user *models.User) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Updates(user).Error
}

func NewRepository(db *gorm.DB) *Repository {
	minioClient, bucket := InitMinio()
	return &Repository{
		db:     db,
		Minio:  minioClient,
		Bucket: bucket,
	}
}

func InitMinio() (*minio.Client, string) {
	endpoint := "127.0.0.1:9000"
	accessKey := "admin"
	secretKey := "admin123"
	bucket := "chrono"

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalf("Ошибка подключения к MinIO: %v", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		log.Fatalf("Ошибка проверки бакета MinIO: %v", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalf("Ошибка создания бакета MinIO: %v", err)
		}
		log.Printf("Бакет %s создан", bucket)
	}

	return client, bucket
}
