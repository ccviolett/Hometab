package repository

import (
	"encoding/json"
	"hometab/internal/model"

	"gorm.io/gorm"
)

type SettingRepo struct {
	db *gorm.DB
}

func NewSettingRepo(db *gorm.DB) *SettingRepo {
	return &SettingRepo{db: db}
}

func (r *SettingRepo) FindAll() ([]model.Setting, error) {
	items := make([]model.Setting, 0)
	err := r.db.Find(&items).Error
	return items, err
}

func (r *SettingRepo) FindAllAsMap() (map[string]interface{}, error) {
	items, err := r.FindAll()
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{}, len(items))
	for _, item := range items {
		m[item.Key] = item.GetValue()
	}
	return m, nil
}

func (r *SettingRepo) FindByKey(key string) (*model.Setting, error) {
	var item model.Setting
	err := r.db.Where("`key` = ?", key).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *SettingRepo) Upsert(key string, value interface{}) (*model.Setting, error) {
	valJSON, _ := json.Marshal(value)
	var existing model.Setting
	err := r.db.Where("`key` = ?", key).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		s := model.Setting{Key: key, ValueJSON: string(valJSON)}
		if err := r.db.Create(&s).Error; err != nil {
			return nil, err
		}
		return &s, nil
	} else if err != nil {
		return nil, err
	}
	existing.ValueJSON = string(valJSON)
	if err := r.db.Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

func (r *SettingRepo) Update(key string, value interface{}) (*model.Setting, error) {
	s, err := r.FindByKey(key)
	if err != nil {
		return nil, err
	}
	valJSON, _ := json.Marshal(value)
	s.ValueJSON = string(valJSON)
	if err := r.db.Save(s).Error; err != nil {
		return nil, err
	}
	return s, nil
}

func (r *SettingRepo) Delete(key string) error {
	return r.db.Where("`key` = ?", key).Delete(&model.Setting{}).Error
}
