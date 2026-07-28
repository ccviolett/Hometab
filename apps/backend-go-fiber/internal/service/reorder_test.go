package service

import (
	"testing"

	"hometab/internal/model"
	"hometab/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReorderScopesAreValidatedAndNormalized(t *testing.T) {
	db := setupTestDB(t)
	groupRepo := repository.NewLinkGroupRepo(db)
	linkRepo := repository.NewLinkRepo(db)
	flowRepo := repository.NewLinkFlowRepo(db)
	itemRepo := repository.NewLinkFlowItemRepo(db)
	groupSvc := NewLinkGroupSvc(groupRepo, linkRepo, flowRepo, itemRepo)
	linkSvc := NewLinkSvc(linkRepo, flowRepo, itemRepo)
	flowSvc := NewLinkFlowSvc(flowRepo, itemRepo, linkRepo)

	groupA := model.LinkGroup{Name: "A"}
	groupB := model.LinkGroup{Name: "B"}
	db.Create(&groupA)
	db.Create(&groupB)
	require.NoError(t, groupSvc.Reorder([]string{groupB.ID.String(), groupA.ID.String()}))
	groups, err := groupRepo.FindAllOrdered()
	require.NoError(t, err)
	assert.Equal(t, groupB.ID, groups[0].ID)
	assert.Equal(t, 0, groups[0].OrderIndex)
	assert.Equal(t, 10, groups[1].OrderIndex)
	assert.Error(t, groupSvc.Reorder([]string{groupA.ID.String()}))

	linkA := model.Link{Name: "A", URL: "https://a", GroupID: &groupA.ID}
	linkB := model.Link{Name: "B", URL: "https://b", GroupID: &groupA.ID}
	db.Create(&linkA)
	db.Create(&linkB)
	require.NoError(t, linkSvc.ReorderGroup(groupA.ID, []string{linkB.ID.String(), linkA.ID.String()}))
	links, err := linkRepo.FindByGroupID(groupA.ID)
	require.NoError(t, err)
	assert.Equal(t, linkB.ID, links[0].ID)

	flow := model.LinkFlow{GroupID: groupA.ID, Name: "F"}
	db.Create(&flow)
	db.Create(&model.LinkFlowItem{FlowID: flow.ID, LinkID: linkA.ID, OrderIndex: 0})
	db.Create(&model.LinkFlowItem{FlowID: flow.ID, LinkID: linkB.ID, OrderIndex: 10})
	require.NoError(t, flowSvc.ReorderLinks(flow.ID, []string{linkB.ID.String(), linkA.ID.String()}))
	itemA, err := itemRepo.Find(flow.ID, linkA.ID)
	require.NoError(t, err)
	itemB, err := itemRepo.Find(flow.ID, linkB.ID)
	require.NoError(t, err)
	assert.Equal(t, 10, itemA.OrderIndex)
	assert.Equal(t, 0, itemB.OrderIndex)
}
