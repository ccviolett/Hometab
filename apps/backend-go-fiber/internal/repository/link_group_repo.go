package repository

import (
	"hometab/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LinkGroupRepo struct {
	db *gorm.DB
}

func NewLinkGroupRepo(db *gorm.DB) *LinkGroupRepo {
	return &LinkGroupRepo{db: db}
}

func (r *LinkGroupRepo) FindAllOrdered() ([]model.LinkGroup, error) {
	items := make([]model.LinkGroup, 0)
	err := r.db.Order("order_index, created_at").Find(&items).Error
	return items, err
}

func (r *LinkGroupRepo) FindByID(id uuid.UUID) (*model.LinkGroup, error) {
	var item model.LinkGroup
	err := r.db.First(&item, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *LinkGroupRepo) FindByName(name string) (*model.LinkGroup, error) {
	var item model.LinkGroup
	err := r.db.Where("name = ?", name).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *LinkGroupRepo) Create(item *model.LinkGroup) error {
	return r.db.Create(item).Error
}

func (r *LinkGroupRepo) Update(item *model.LinkGroup) error {
	return r.db.Save(item).Error
}

func (r *LinkGroupRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.LinkGroup{}, "id = ?", id).Error
}

func (r *LinkGroupRepo) Reorder(ids []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for index, id := range ids {
			result := tx.Model(&model.LinkGroup{}).Where("id = ?", id).Update("order_index", index*10)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return gorm.ErrRecordNotFound
			}
		}
		return nil
	})
}
