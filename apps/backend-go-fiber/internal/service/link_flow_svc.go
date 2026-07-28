package service

import (
	"errors"
	"hometab/internal/model"
	"hometab/internal/repository"

	"github.com/google/uuid"
)

type LinkFlowSvc struct {
	repo     *repository.LinkFlowRepo
	itemRepo *repository.LinkFlowItemRepo
	linkRepo *repository.LinkRepo
}

func NewLinkFlowSvc(repo *repository.LinkFlowRepo, itemRepo *repository.LinkFlowItemRepo, linkRepo *repository.LinkRepo) *LinkFlowSvc {
	return &LinkFlowSvc{repo: repo, itemRepo: itemRepo, linkRepo: linkRepo}
}

func (s *LinkFlowSvc) FindAll(groupID *string) ([]model.LinkFlow, error) {
	if groupID != nil && *groupID != "" {
		gid, err := uuid.Parse(*groupID)
		if err != nil {
			return nil, errors.New("invalid group_id")
		}
		return s.repo.FindByGroupID(gid)
	}
	return s.repo.FindAll()
}

func (s *LinkFlowSvc) FindByID(id uuid.UUID) (*model.LinkFlow, error) {
	return s.repo.FindByID(id)
}

func (s *LinkFlowSvc) Create(req model.LinkFlowCreate) (*model.LinkFlow, error) {
	gid, err := uuid.Parse(req.GroupID)
	if err != nil {
		return nil, errors.New("invalid group_id")
	}
	orderIdx := 0
	if req.OrderIndex != nil {
		orderIdx = *req.OrderIndex
	} else {
		max, _ := s.repo.MaxOrderIndex(gid)
		orderIdx = max + 1
	}
	flow := model.LinkFlow{
		GroupID:    gid,
		Name:       req.Name,
		OrderIndex: orderIdx,
	}
	if err := s.repo.Create(&flow); err != nil {
		return nil, err
	}
	return &flow, nil
}

func (s *LinkFlowSvc) Update(id uuid.UUID, req model.LinkFlowUpdate) (*model.LinkFlow, error) {
	flow, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		flow.Name = *req.Name
	}
	if req.OrderIndex != nil {
		flow.OrderIndex = *req.OrderIndex
	}
	if req.GroupID != nil {
		gid, err := uuid.Parse(*req.GroupID)
		if err != nil {
			return nil, errors.New("invalid group_id")
		}
		flow.GroupID = gid
	}
	if err := s.repo.Update(flow); err != nil {
		return nil, err
	}
	return flow, nil
}

func (s *LinkFlowSvc) Delete(id uuid.UUID, _ *model.LinkFlowDeleteOptions) error {
	if err := s.itemRepo.DeleteByFlowID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func (s *LinkFlowSvc) AddLink(flowID uuid.UUID, req model.LinkFlowLinkRequest) (*model.Link, error) {
	flow, err := s.repo.FindByID(flowID)
	if err != nil {
		return nil, errors.New("flow not found")
	}
	linkID, err := uuid.Parse(req.LinkID)
	if err != nil {
		return nil, errors.New("invalid link_id")
	}
	link, err := s.linkRepo.FindByID(linkID)
	if err != nil {
		return nil, errors.New("link not found")
	}
	if link.GroupID != nil && *link.GroupID != flow.GroupID {
		return nil, errors.New("link group_id does not match flow group_id")
	}
	orderIndex := link.OrderIndex
	if req.OrderIndex != nil {
		orderIndex = *req.OrderIndex
	}
	if err := s.itemRepo.Add(&model.LinkFlowItem{
		FlowID:     flowID,
		LinkID:     link.ID,
		OrderIndex: orderIndex,
	}); err != nil {
		return nil, err
	}
	return link, nil
}

func (s *LinkFlowSvc) UpdateLinkOrder(flowID, linkID uuid.UUID, req model.LinkFlowLinkOrderUpdate) (*model.Link, error) {
	link, err := s.linkRepo.FindByID(linkID)
	if err != nil {
		return nil, errors.New("link not found")
	}
	if err := s.itemRepo.UpdateOrder(flowID, linkID, req.OrderIndex); err != nil {
		return nil, errors.New("link does not belong to this flow")
	}
	return link, nil
}

func (s *LinkFlowSvc) RemoveLink(flowID, linkID uuid.UUID) error {
	if _, err := s.linkRepo.FindByID(linkID); err != nil {
		return errors.New("link not found")
	}
	if err := s.itemRepo.Remove(flowID, linkID); err != nil {
		return errors.New("link does not belong to this flow")
	}
	return nil
}
