package service

import (
	"hometab/internal/model"
	"hometab/internal/repository"
)

type SettingSvc struct {
	repo *repository.SettingRepo
}

func NewSettingSvc(repo *repository.SettingRepo) *SettingSvc {
	return &SettingSvc{repo: repo}
}

func (s *SettingSvc) FindAll() (map[string]interface{}, error) {
	return s.repo.FindAllAsMap()
}

func (s *SettingSvc) FindByKey(key string) (*model.SettingRead, error) {
	item, err := s.repo.FindByKey(key)
	if err != nil {
		return nil, err
	}
	return &model.SettingRead{Key: item.Key, Value: item.GetValue()}, nil
}

func (s *SettingSvc) CreateOrUpdate(req model.SettingCreate) (*model.SettingRead, error) {
	item, err := s.repo.Upsert(req.Key, req.Value)
	if err != nil {
		return nil, err
	}
	return &model.SettingRead{Key: item.Key, Value: item.GetValue()}, nil
}

func (s *SettingSvc) Update(key string, req model.SettingUpdate) (*model.SettingRead, error) {
	item, err := s.repo.Update(key, req.Value)
	if err != nil {
		return nil, err
	}
	return &model.SettingRead{Key: item.Key, Value: item.GetValue()}, nil
}

func (s *SettingSvc) Delete(key string) error {
	return s.repo.Delete(key)
}
