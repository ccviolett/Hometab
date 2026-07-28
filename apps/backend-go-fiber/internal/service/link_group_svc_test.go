package service

import (
	"testing"

	"hometab/internal/model"
	"hometab/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkGroupSvcCRUD(t *testing.T) {
	db := setupTestDB(t)
	groupRepo := repository.NewLinkGroupRepo(db)
	linkRepo := repository.NewLinkRepo(db)
	flowRepo := repository.NewLinkFlowRepo(db)
	svc := NewLinkGroupSvc(groupRepo, linkRepo, flowRepo, repository.NewLinkFlowItemRepo(db))

	// FindAll empty
	items, err := svc.FindAll()
	require.NoError(t, err)
	assert.Len(t, items, 0)

	// Create
	desc := "desc"
	created, err := svc.Create(model.LinkGroupCreate{Name: "G1", Description: &desc, OrderIndex: 1})
	require.NoError(t, err)
	assert.Equal(t, "G1", created.Name)
	assert.Equal(t, &desc, created.Description)

	// FindByID
	got, err := svc.FindByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "G1", got.Name)

	// FindByID not found
	_, err = svc.FindByID(uuid.New())
	assert.Error(t, err)

	// Update
	newName := "Renamed"
	newDesc := "new desc"
	newIdx := 2
	updated, err := svc.Update(created.ID, model.LinkGroupUpdate{
		Name: &newName, Description: &newDesc, OrderIndex: &newIdx,
	})
	require.NoError(t, err)
	assert.Equal(t, "Renamed", updated.Name)
	assert.Equal(t, &newDesc, updated.Description)
	assert.Equal(t, 2, updated.OrderIndex)

	// Update not found
	_, err = svc.Update(uuid.New(), model.LinkGroupUpdate{Name: &newName})
	assert.Error(t, err)

	// Delete (creates "未分组", moves links)
	link := model.Link{Name: "L1", URL: "https://a.com", GroupID: &created.ID}
	db.Create(&link)
	flow := model.LinkFlow{GroupID: created.ID, Name: "F1"}
	db.Create(&flow)

	err = svc.Delete(created.ID)
	assert.NoError(t, err)

	// Verify "未分组" was created
	var ungrouped model.LinkGroup
	db.Where("name = ?", "未分组").First(&ungrouped)
	assert.Equal(t, "未分组", ungrouped.Name)

	// Verify link moved to 未分组
	var movedLink model.Link
	db.First(&movedLink, "id = ?", link.ID)
	assert.Equal(t, ungrouped.ID, *movedLink.GroupID)
}

func TestLinkGroupSvcDeleteWithExistingUngrouped(t *testing.T) {
	db := setupTestDB(t)
	groupRepo := repository.NewLinkGroupRepo(db)
	linkRepo := repository.NewLinkRepo(db)
	flowRepo := repository.NewLinkFlowRepo(db)
	svc := NewLinkGroupSvc(groupRepo, linkRepo, flowRepo, repository.NewLinkFlowItemRepo(db))

	// Pre-create "未分组"
	ungrouped := model.LinkGroup{Name: "未分组", OrderIndex: 9999}
	db.Create(&ungrouped)

	// Create and delete another group
	g, _ := svc.Create(model.LinkGroupCreate{Name: "ToDelete"})
	err := svc.Delete(g.ID)
	assert.NoError(t, err)
}

func TestLinkGroupSvcListByGroup(t *testing.T) {
	db := setupTestDB(t)
	groupRepo := repository.NewLinkGroupRepo(db)
	linkRepo := repository.NewLinkRepo(db)
	flowRepo := repository.NewLinkFlowRepo(db)
	svc := NewLinkGroupSvc(groupRepo, linkRepo, flowRepo, repository.NewLinkFlowItemRepo(db))

	// Empty result
	result, err := svc.ListByGroup()
	require.NoError(t, err)
	assert.Len(t, result, 0)

	// Create group with links and flows
	g := model.LinkGroup{Name: "G1"}
	db.Create(&g)

	flow := model.LinkFlow{GroupID: g.ID, Name: "F1"}
	db.Create(&flow)

	// A flow references an existing grouped link without removing it from the group.
	link1 := model.Link{Name: "L1", URL: "https://a.com", GroupID: &g.ID}
	db.Create(&link1)
	db.Create(&model.LinkFlowItem{FlowID: flow.ID, LinkID: link1.ID})

	// Ungrouped link (no flow)
	link2 := model.Link{Name: "L2", URL: "https://b.com", GroupID: &g.ID}
	db.Create(&link2)

	result, err = svc.ListByGroup()
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "G1", result[0].Group.Name)
	assert.Len(t, result[0].Flows, 1)
	assert.Len(t, result[0].Flows[0].Links, 1)
	assert.Len(t, result[0].Links, 2)
}
