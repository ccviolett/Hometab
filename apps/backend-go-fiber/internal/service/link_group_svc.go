package service

import (
	"hometab/internal/model"
	"hometab/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LinkGroupSvc struct {
	repo         *repository.LinkGroupRepo
	linkRepo     *repository.LinkRepo
	flowRepo     *repository.LinkFlowRepo
	flowItemRepo *repository.LinkFlowItemRepo
}

func NewLinkGroupSvc(repo *repository.LinkGroupRepo, linkRepo *repository.LinkRepo, flowRepo *repository.LinkFlowRepo, flowItemRepo *repository.LinkFlowItemRepo) *LinkGroupSvc {
	return &LinkGroupSvc{repo: repo, linkRepo: linkRepo, flowRepo: flowRepo, flowItemRepo: flowItemRepo}
}

func (s *LinkGroupSvc) FindAll() ([]model.LinkGroup, error) {
	return s.repo.FindAllOrdered()
}

func (s *LinkGroupSvc) FindByID(id uuid.UUID) (*model.LinkGroup, error) {
	return s.repo.FindByID(id)
}

func (s *LinkGroupSvc) Create(req model.LinkGroupCreate) (*model.LinkGroup, error) {
	g := model.LinkGroup{
		Name:        req.Name,
		Description: req.Description,
		OrderIndex:  req.OrderIndex,
	}
	if err := s.repo.Create(&g); err != nil {
		return nil, err
	}
	return &g, nil
}

func (s *LinkGroupSvc) Update(id uuid.UUID, req model.LinkGroupUpdate) (*model.LinkGroup, error) {
	g, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		g.Name = *req.Name
	}
	if req.Description != nil {
		g.Description = req.Description
	}
	if req.OrderIndex != nil {
		g.OrderIndex = *req.OrderIndex
	}
	if err := s.repo.Update(g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *LinkGroupSvc) Delete(id uuid.UUID) error {
	ungrouped, err := s.repo.FindByName("未分组")
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ug := model.LinkGroup{Name: "未分组", OrderIndex: 9999}
			if err := s.repo.Create(&ug); err != nil {
				return err
			}
			ungrouped = &ug
		} else {
			return err
		}
	}
	if err := s.linkRepo.MoveToGroup(id, ungrouped.ID); err != nil {
		return err
	}
	if err := s.flowRepo.DeleteByGroupID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func (s *LinkGroupSvc) ListByGroup() ([]model.GroupedLinksResponse, error) {
	groups, err := s.repo.FindAllOrdered()
	if err != nil {
		return nil, err
	}
	result := make([]model.GroupedLinksResponse, 0, len(groups))
	for _, g := range groups {
		flows, err := s.flowRepo.FindByGroupID(g.ID)
		if err != nil {
			return nil, err
		}
		flowsWithLinks := make([]model.LinkFlowWithLinks, 0, len(flows))
		for _, f := range flows {
			links, err := s.flowItemRepo.FindLinks(f.ID)
			if err != nil {
				return nil, err
			}
			if links == nil {
				links = []model.Link{}
			}
			flowsWithLinks = append(flowsWithLinks, model.LinkFlowWithLinks{Flow: f, Links: links})
		}
		ungroupedLinks, err := s.linkRepo.FindByGroupID(g.ID)
		if err != nil {
			return nil, err
		}
		if ungroupedLinks == nil {
			ungroupedLinks = []model.Link{}
		}
		result = append(result, model.GroupedLinksResponse{
			Group: g,
			Flows: flowsWithLinks,
			Links: ungroupedLinks,
		})
	}
	return result, nil
}
