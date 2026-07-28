package repository

import (
	"hometab/internal/model"

	"gorm.io/gorm"
)

type DomainIconRepo struct {
	db *gorm.DB
}

func NewDomainIconRepo(db *gorm.DB) *DomainIconRepo {
	return &DomainIconRepo{db: db}
}

func (r *DomainIconRepo) FindAll() ([]model.DomainIcon, error) {
	items := make([]model.DomainIcon, 0)
	err := r.db.Order("host").Find(&items).Error
	return items, err
}

func (r *DomainIconRepo) FindByHost(host string) (*model.DomainIcon, error) {
	var item model.DomainIcon
	err := r.db.First(&item, "host = ?", host).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *DomainIconRepo) Save(item *model.DomainIcon) error {
	return r.db.Save(item).Error
}

func (r *DomainIconRepo) Delete(host string) error {
	return r.db.Delete(&model.DomainIcon{}, "host = ?", host).Error
}
