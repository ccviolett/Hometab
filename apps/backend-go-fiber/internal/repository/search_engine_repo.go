package repository

import (
	"hometab/internal/model"

	"gorm.io/gorm"
)

type SearchEngineRepo struct {
	db *gorm.DB
}

func NewSearchEngineRepo(db *gorm.DB) *SearchEngineRepo {
	return &SearchEngineRepo{db: db}
}

func (r *SearchEngineRepo) FindAll() ([]model.SearchEngine, error) {
	items := make([]model.SearchEngine, 0)
	err := r.db.Order("id").Find(&items).Error
	return items, err
}

func (r *SearchEngineRepo) FindByID(id uint) (*model.SearchEngine, error) {
	var item model.SearchEngine
	err := r.db.First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *SearchEngineRepo) Create(item *model.SearchEngine) error {
	return r.db.Create(item).Error
}

func (r *SearchEngineRepo) Update(item *model.SearchEngine) error {
	return r.db.Save(item).Error
}

func (r *SearchEngineRepo) Delete(id uint) error {
	return r.db.Delete(&model.SearchEngine{}, id).Error
}
