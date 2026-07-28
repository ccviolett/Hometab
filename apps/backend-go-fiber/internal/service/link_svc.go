package service

import (
	"errors"
	"hometab/internal/model"
	"hometab/internal/repository"

	"github.com/google/uuid"
)

type LinkSvc struct {
	repo         *repository.LinkRepo
	flowRepo     *repository.LinkFlowRepo
	flowItemRepo *repository.LinkFlowItemRepo
}

func NewLinkSvc(repo *repository.LinkRepo, flowRepo *repository.LinkFlowRepo, flowItemRepo *repository.LinkFlowItemRepo) *LinkSvc {
	return &LinkSvc{repo: repo, flowRepo: flowRepo, flowItemRepo: flowItemRepo}
}

func (s *LinkSvc) FindAll() ([]model.Link, error) {
	return s.repo.FindAll()
}

func (s *LinkSvc) FindByID(id uuid.UUID) (*model.Link, error) {
	return s.repo.FindByID(id)
}

func (s *LinkSvc) Create(req model.LinkCreate) (*model.Link, error) {
	link := model.Link{
		Name:       req.Name,
		URL:        req.URL,
		OrderIndex: req.OrderIndex,
	}
	if req.GroupID != nil {
		gid, err := uuid.Parse(*req.GroupID)
		if err != nil {
			return nil, errors.New("invalid group_id")
		}
		link.GroupID = &gid
	}
	if req.FlowID != nil {
		fid, err := uuid.Parse(*req.FlowID)
		if err != nil {
			return nil, errors.New("invalid flow_id")
		}
		flow, err := s.flowRepo.FindByID(fid)
		if err != nil {
			return nil, errors.New("flow not found")
		}
		if link.GroupID == nil {
			link.GroupID = &flow.GroupID
		} else if *link.GroupID != flow.GroupID {
			return nil, errors.New("link group_id does not match flow group_id")
		}
	}
	if err := s.repo.Create(&link); err != nil {
		return nil, err
	}
	if req.FlowID != nil {
		fid, _ := uuid.Parse(*req.FlowID)
		if err := s.flowItemRepo.Add(&model.LinkFlowItem{FlowID: fid, LinkID: link.ID, OrderIndex: link.OrderIndex}); err != nil {
			_ = s.repo.Delete(link.ID)
			return nil, err
		}
	}
	return &link, nil
}

func (s *LinkSvc) Update(id uuid.UUID, req model.LinkUpdate) (*model.Link, error) {
	link, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		link.Name = *req.Name
	}
	if req.URL != nil {
		link.URL = *req.URL
	}
	if req.OrderIndex != nil {
		link.OrderIndex = *req.OrderIndex
	}
	if req.GroupID != nil {
		gid, err := uuid.Parse(*req.GroupID)
		if err != nil {
			return nil, errors.New("invalid group_id")
		}
		link.GroupID = &gid
	}
	var addToFlow *uuid.UUID
	if req.FlowID != nil && *req.FlowID != "" {
		fid, err := uuid.Parse(*req.FlowID)
		if err != nil {
			return nil, errors.New("invalid flow_id")
		}
		flow, err := s.flowRepo.FindByID(fid)
		if err != nil {
			return nil, errors.New("flow not found")
		}
		if link.GroupID != nil && *link.GroupID != flow.GroupID {
			return nil, errors.New("link group_id does not match flow group_id")
		}
		addToFlow = &fid
	}
	if err := s.repo.Update(link); err != nil {
		return nil, err
	}
	if addToFlow != nil {
		if _, err := s.flowItemRepo.Find(*addToFlow, link.ID); err != nil {
			if err := s.flowItemRepo.Add(&model.LinkFlowItem{FlowID: *addToFlow, LinkID: link.ID, OrderIndex: link.OrderIndex}); err != nil {
				return nil, err
			}
		}
	}
	return link, nil
}

func (s *LinkSvc) Delete(id uuid.UUID) error {
	if err := s.flowItemRepo.DeleteByLinkID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
