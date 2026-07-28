package service

import (
	"hometab/internal/model"
	"hometab/internal/repository"
)

type SearchEngineSvc struct {
	repo *repository.SearchEngineRepo
}

func NewSearchEngineSvc(repo *repository.SearchEngineRepo) *SearchEngineSvc {
	return &SearchEngineSvc{repo: repo}
}

func (s *SearchEngineSvc) FindAll() ([]model.SearchEngine, error) {
	return s.repo.FindAll()
}

func (s *SearchEngineSvc) FindByID(id uint) (*model.SearchEngine, error) {
	return s.repo.FindByID(id)
}

func (s *SearchEngineSvc) Create(req model.SearchEngineCreate) (*model.SearchEngine, error) {
	e := model.SearchEngine{
		Name:        req.Name,
		URLTemplate: req.URLTemplate,
		Icon:        req.Icon,
		Description: req.Description,
		Color:       req.Color,
	}
	if err := s.repo.Create(&e); err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *SearchEngineSvc) Update(id uint, req model.SearchEngineUpdate) (*model.SearchEngine, error) {
	e, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		e.Name = *req.Name
	}
	if req.URLTemplate != nil {
		e.URLTemplate = *req.URLTemplate
	}
	if req.Icon != nil {
		e.Icon = req.Icon
	}
	if req.Description != nil {
		e.Description = req.Description
	}
	if req.Color != nil {
		e.Color = req.Color
	}
	if err := s.repo.Update(e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *SearchEngineSvc) Delete(id uint) error {
	return s.repo.Delete(id)
}
