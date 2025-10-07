package repository

import (
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"rip/internal/app/models"
)

type LayersRepository struct {
	db *sql.DB
}

func NewLayersRepository(db *sql.DB) *LayersRepository {
	return &LayersRepository{db: db}
}

func (r *LayersRepository) GetLayers() ([]models.Layer, error) {
	rows, err := r.db.Query("SELECT id, name, description, image_url, from_year, to_year, status FROM layers ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var layers []models.Layer
	for rows.Next() {
		var l models.Layer
		if err := rows.Scan(&l.ID, &l.Name, &l.Description, &l.ImageURL, &l.FromYear, &l.ToYear, &l.Status); err != nil {
			return nil, err
		}
		layers = append(layers, l)
	}
	return layers, nil
}

func (r *LayersRepository) GetLayerByID(id uint) (*models.Layer, error) {
	var l models.Layer
	err := r.db.QueryRow("SELECT id, name, description, image_url, from_year, to_year, status FROM layers WHERE id=$1", id).
		Scan(&l.ID, &l.Name, &l.Description, &l.ImageURL, &l.FromYear, &l.ToYear, &l.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *LayersRepository) CreateLayer(l *models.Layer) error {
	err := r.db.QueryRow(
		`INSERT INTO layers (name, description, image_url, from_year, to_year, status) 
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		l.Name, l.Description, l.ImageURL, l.FromYear, l.ToYear, l.Status,
	).Scan(&l.ID)
	return err
}

func (r *LayersRepository) UpdateLayer(id uint, l *models.Layer) error {
	_, err := r.db.Exec(
		`UPDATE layers SET name=$1, description=$2, image_url=$3, from_year=$4, to_year=$5, status=$6 WHERE id=$7`,
		l.Name, l.Description, l.ImageURL, l.FromYear, l.ToYear, l.Status, id,
	)
	return err
}

func (r *LayersRepository) DeleteLayer(id uint) error {
	_, err := r.db.Exec("DELETE FROM layers WHERE id=$1", id)
	return err
}

func (r *LayersRepository) SaveLayerImage(id uint, fileHeader *multipart.FileHeader) error {
	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dir := "static/images/layers"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	dstPath := filepath.Join(dir, fmt.Sprintf("%d_%s", id, fileHeader.Filename))
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	imageURL := "/" + dstPath
	_, err = r.db.Exec("UPDATE layers SET image_url=$1 WHERE id=$2", imageURL, id)
	return err
}

func (r *Repository) CreateLayer(layer *models.Layer) error {
	return r.db.Create(layer).Error
}

func (r *Repository) UpdateLayer(id uint, input *models.Layer) error {
	return r.db.Model(&models.Layer{}).Where("id = ?", id).Updates(input).Error
}

func (r *Repository) DeleteLayer(id uint) error {
	return r.db.Delete(&models.Layer{}, id).Error
}
