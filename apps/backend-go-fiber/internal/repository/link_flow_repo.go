package repository

import (
	"hometab/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LinkFlowRepo struct {
	db *gorm.DB
}

func NewLinkFlowRepo(db *gorm.DB) *LinkFlowRepo {
	return &LinkFlowRepo{db: db}
}

func (r *LinkFlowRepo) FindAll() ([]model.LinkFlow, error) {
	items := make([]model.LinkFlow, 0)
	err := r.db.Order("order_index, created_at").Find(&items).Error
	return items, err
}

func (r *LinkFlowRepo) FindByGroupID(gid uuid.UUID) ([]model.LinkFlow, error) {
	items := make([]model.LinkFlow, 0)
	err := r.db.Where("group_id = ?", gid).Order("order_index, created_at").Find(&items).Error
	return items, err
}

func (r *LinkFlowRepo) FindByID(id uuid.UUID) (*model.LinkFlow, error) {
	var item model.LinkFlow
	err := r.db.First(&item, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *LinkFlowRepo) MaxOrderIndex(gid uuid.UUID) (int, error) {
	var max *int
	err := r.db.Model(&model.LinkFlow{}).Where("group_id = ?", gid).Select("COALESCE(MAX(order_index), -1)").Scan(&max).Error
	if err != nil || max == nil {
		return -1, err
	}
	return *max, nil
}

func (r *LinkFlowRepo) Create(item *model.LinkFlow) error {
	return r.db.Create(item).Error
}

func (r *LinkFlowRepo) Update(item *model.LinkFlow) error {
	return r.db.Save(item).Error
}

func (r *LinkFlowRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.LinkFlow{}, "id = ?", id).Error
}

func (r *LinkFlowRepo) DeleteByGroupID(gid uuid.UUID) error {
	return r.db.Where("group_id = ?", gid).Delete(&model.LinkFlow{}).Error
}
