package service

import (
	"errors"

	"github.com/google/uuid"
)

func parseReorderIDs(raw []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(raw))
	seen := make(map[uuid.UUID]bool)
	for _, value := range raw {
		id, err := uuid.Parse(value)
		if err != nil || seen[id] {
			return nil, errors.New("ids must contain unique valid UUIDs")
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids, nil
}

func validateReorderSet(expected []uuid.UUID, actual []uuid.UUID) error {
	if len(expected) != len(actual) {
		return errors.New("ids must contain the complete scope")
	}
	set := make(map[uuid.UUID]bool, len(expected))
	for _, id := range expected {
		set[id] = true
	}
	for _, id := range actual {
		if !set[id] {
			return errors.New("ids contain an item outside the scope")
		}
	}
	return nil
}

func (s *LinkGroupSvc) Reorder(rawIDs []string) error {
	ids, err := parseReorderIDs(rawIDs)
	if err != nil {
		return err
	}
	groups, err := s.repo.FindAllOrdered()
	if err != nil {
		return err
	}
	expected := make([]uuid.UUID, 0, len(groups))
	for _, group := range groups {
		expected = append(expected, group.ID)
	}
	if err := validateReorderSet(expected, ids); err != nil {
		return err
	}
	return s.repo.Reorder(ids)
}

func (s *LinkSvc) ReorderGroup(groupID uuid.UUID, rawIDs []string) error {
	ids, err := parseReorderIDs(rawIDs)
	if err != nil {
		return err
	}
	links, err := s.repo.FindByGroupID(groupID)
	if err != nil {
		return err
	}
	expected := make([]uuid.UUID, 0, len(links))
	for _, link := range links {
		expected = append(expected, link.ID)
	}
	if err := validateReorderSet(expected, ids); err != nil {
		return err
	}
	return s.repo.ReorderGroup(groupID, ids)
}

func (s *LinkFlowSvc) ReorderLinks(flowID uuid.UUID, rawIDs []string) error {
	ids, err := parseReorderIDs(rawIDs)
	if err != nil {
		return err
	}
	links, err := s.itemRepo.FindLinks(flowID)
	if err != nil {
		return err
	}
	expected := make([]uuid.UUID, 0, len(links))
	for _, link := range links {
		expected = append(expected, link.ID)
	}
	if err := validateReorderSet(expected, ids); err != nil {
		return err
	}
	return s.itemRepo.Reorder(flowID, ids)
}
