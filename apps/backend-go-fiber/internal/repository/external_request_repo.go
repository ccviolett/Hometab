package repository

import (
	"hometab/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ExternalRequestRepo struct {
	db *gorm.DB
}

func NewExternalRequestRepo(db *gorm.DB) *ExternalRequestRepo {
	return &ExternalRequestRepo{db: db}
}

func (r *ExternalRequestRepo) FindAll() ([]model.ExternalRequest, error) {
	items := make([]model.ExternalRequest, 0)
	err := r.db.Order("order_index, created_at").Find(&items).Error
	return items, err
}

func (r *ExternalRequestRepo) FindByID(id uuid.UUID) (*model.ExternalRequest, error) {
	var item model.ExternalRequest
	err := r.db.First(&item, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ExternalRequestRepo) Create(item *model.ExternalRequest) error {
	return r.db.Create(item).Error
}

func (r *ExternalRequestRepo) Update(item *model.ExternalRequest) error {
	return r.db.Save(item).Error
}

func (r *ExternalRequestRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.ExternalRequest{}, "id = ?", id).Error
}
