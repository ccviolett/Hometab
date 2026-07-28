package service

import (
	"testing"

	"hometab/internal/model"
	"hometab/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkFlowSvcCRUD(t *testing.T) {
	db := setupTestDB(t)
	flowRepo := repository.NewLinkFlowRepo(db)
	linkRepo := repository.NewLinkRepo(db)
	svc := NewLinkFlowSvc(flowRepo, repository.NewLinkFlowItemRepo(db), linkRepo)

	g := model.LinkGroup{Name: "G1"}
	db.Create(&g)
	gid := g.ID.String()

	// Create flow (auto order index)
	flow, err := svc.Create(model.LinkFlowCreate{GroupID: gid, Name: "F1"})
	require.NoError(t, err)
	assert.Equal(t, "F1", flow.Name)
	assert.Equal(t, 0, flow.OrderIndex)

	// Create flow with explicit order index
	idx := 5
	flow2, err := svc.Create(model.LinkFlowCreate{GroupID: gid, Name: "F2", OrderIndex: &idx})
	require.NoError(t, err)
	assert.Equal(t, 5, flow2.OrderIndex)

	// Create with invalid group_id
	_, err = svc.Create(model.LinkFlowCreate{GroupID: "bad", Name: "F3"})
	assert.Error(t, err)

	// FindByID
	got, err := svc.FindByID(flow.ID)
	require.NoError(t, err)
	assert.Equal(t, "F1", got.Name)

	// FindAll
	items, err := svc.FindAll(nil)
	require.NoError(t, err)
	assert.Len(t, items, 2)

	// FindAll with group filter
	items, err = svc.FindAll(&gid)
	require.NoError(t, err)
	assert.Len(t, items, 2)

	// FindAll with invalid group filter
	badGid := "not-uuid"
	_, err = svc.FindAll(&badGid)
	assert.Error(t, err)

	// Update name
	newName := "Renamed"
	updated, err := svc.Update(flow.ID, model.LinkFlowUpdate{Name: &newName})
	require.NoError(t, err)
	assert.Equal(t, "Renamed", updated.Name)

	// Update order
	newIdx := 10
	updated, err = svc.Update(flow.ID, model.LinkFlowUpdate{OrderIndex: &newIdx})
	require.NoError(t, err)
	assert.Equal(t, 10, updated.OrderIndex)

	// Update group_id
	g2 := model.LinkGroup{Name: "G2"}
	db.Create(&g2)
	gid2 := g2.ID.String()
	updated, err = svc.Update(flow.ID, model.LinkFlowUpdate{GroupID: &gid2})
	require.NoError(t, err)
	assert.Equal(t, g2.ID, updated.GroupID)

	// Update with invalid group_id
	badG := "bad"
	_, err = svc.Update(flow.ID, model.LinkFlowUpdate{GroupID: &badG})
	assert.Error(t, err)

	// Update not found
	_, err = svc.Update(uuid.New(), model.LinkFlowUpdate{Name: &newName})
	assert.Error(t, err)

	// Delete
	err = svc.Delete(flow2.ID, nil)
	assert.NoError(t, err)
}

func TestLinkFlowSvcDeletePreservesLinks(t *testing.T) {
	db := setupTestDB(t)
	flowRepo := repository.NewLinkFlowRepo(db)
	itemRepo := repository.NewLinkFlowItemRepo(db)
	linkRepo := repository.NewLinkRepo(db)
	svc := NewLinkFlowSvc(flowRepo, itemRepo, linkRepo)

	g := model.LinkGroup{Name: "G1"}
	db.Create(&g)
	flow := model.LinkFlow{GroupID: g.ID, Name: "F1"}
	db.Create(&flow)
	link1 := model.Link{Name: "L1", URL: "https://a.com", GroupID: &g.ID}
	link2 := model.Link{Name: "L2", URL: "https://b.com", GroupID: &g.ID}
	db.Create(&link1)
	db.Create(&link2)
	require.NoError(t, itemRepo.Add(&model.LinkFlowItem{FlowID: flow.ID, LinkID: link1.ID}))
	require.NoError(t, itemRepo.Add(&model.LinkFlowItem{FlowID: flow.ID, LinkID: link2.ID}))

	require.NoError(t, svc.Delete(flow.ID, nil))

	var linkCount, itemCount int64
	db.Model(&model.Link{}).Where("id IN ?", []uuid.UUID{link1.ID, link2.ID}).Count(&linkCount)
	db.Model(&model.LinkFlowItem{}).Where("flow_id = ?", flow.ID).Count(&itemCount)
	assert.Equal(t, int64(2), linkCount)
	assert.Equal(t, int64(0), itemCount)
}

func TestLinkFlowSvcDeleteWithInvalidKeepIDs(t *testing.T) {
	db := setupTestDB(t)
	flowRepo := repository.NewLinkFlowRepo(db)
	linkRepo := repository.NewLinkRepo(db)
	svc := NewLinkFlowSvc(flowRepo, repository.NewLinkFlowItemRepo(db), linkRepo)

	g := model.LinkGroup{Name: "G1"}
	db.Create(&g)
	flow := model.LinkFlow{GroupID: g.ID, Name: "F1"}
	db.Create(&flow)

	// Delete with invalid keep ID (should be silently ignored)
	opts := &model.LinkFlowDeleteOptions{LinkIDsToKeep: []string{"not-a-uuid"}}
	err := svc.Delete(flow.ID, opts)
	assert.NoError(t, err)
}

func TestLinkFlowSvcAddLink(t *testing.T) {
	db := setupTestDB(t)
	flowRepo := repository.NewLinkFlowRepo(db)
	linkRepo := repository.NewLinkRepo(db)
	svc := NewLinkFlowSvc(flowRepo, repository.NewLinkFlowItemRepo(db), linkRepo)

	g := model.LinkGroup{Name: "G1"}
	db.Create(&g)
	flow := model.LinkFlow{GroupID: g.ID, Name: "F1"}
	db.Create(&flow)
	link := model.Link{Name: "L1", URL: "https://a.com", GroupID: &g.ID}
	db.Create(&link)

	// Add link without changing its group order.
	idx := 3
	result, err := svc.AddLink(flow.ID, model.LinkFlowLinkRequest{LinkID: link.ID.String(), OrderIndex: &idx})
	require.NoError(t, err)
	assert.Nil(t, result.FlowID)
	assert.Equal(t, 0, result.OrderIndex)
	membership, err := repository.NewLinkFlowItemRepo(db).Find(flow.ID, link.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, membership.OrderIndex)

	// Add with invalid flow_id
	_, err = svc.AddLink(uuid.New(), model.LinkFlowLinkRequest{LinkID: link.ID.String()})
	assert.Error(t, err)

	// Add with invalid link_id
	_, err = svc.AddLink(flow.ID, model.LinkFlowLinkRequest{LinkID: "bad"})
	assert.Error(t, err)

	// Add with nonexistent link_id
	_, err = svc.AddLink(flow.ID, model.LinkFlowLinkRequest{LinkID: uuid.New().String()})
	assert.Error(t, err)

	// Add link from different group
	g2 := model.LinkGroup{Name: "G2"}
	db.Create(&g2)
	link2 := model.Link{Name: "L2", URL: "https://b.com", GroupID: &g2.ID}
	db.Create(&link2)
	_, err = svc.AddLink(flow.ID, model.LinkFlowLinkRequest{LinkID: link2.ID.String()})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not match")
}

func TestLinkFlowSvcUpdateLinkOrder(t *testing.T) {
	db := setupTestDB(t)
	flowRepo := repository.NewLinkFlowRepo(db)
	linkRepo := repository.NewLinkRepo(db)
	svc := NewLinkFlowSvc(flowRepo, repository.NewLinkFlowItemRepo(db), linkRepo)

	g := model.LinkGroup{Name: "G1"}
	db.Create(&g)
	flow := model.LinkFlow{GroupID: g.ID, Name: "F1"}
	db.Create(&flow)
	link := model.Link{Name: "L1", URL: "https://a.com", GroupID: &g.ID, OrderIndex: 2}
	db.Create(&link)
	itemRepo := repository.NewLinkFlowItemRepo(db)
	require.NoError(t, itemRepo.Add(&model.LinkFlowItem{FlowID: flow.ID, LinkID: link.ID, OrderIndex: 3}))

	// Update only the flow membership order.
	result, err := svc.UpdateLinkOrder(flow.ID, link.ID, model.LinkFlowLinkOrderUpdate{OrderIndex: 7})
	require.NoError(t, err)
	assert.Equal(t, 2, result.OrderIndex)
	membership, err := itemRepo.Find(flow.ID, link.ID)
	require.NoError(t, err)
	assert.Equal(t, 7, membership.OrderIndex)

	// Link not in flow
	link2 := model.Link{Name: "L2", URL: "https://b.com", GroupID: &g.ID}
	db.Create(&link2)
	_, err = svc.UpdateLinkOrder(flow.ID, link2.ID, model.LinkFlowLinkOrderUpdate{OrderIndex: 1})
	assert.Error(t, err)

	// Nonexistent link
	_, err = svc.UpdateLinkOrder(flow.ID, uuid.New(), model.LinkFlowLinkOrderUpdate{OrderIndex: 1})
	assert.Error(t, err)
}

func TestLinkFlowSvcSupportsIndependentMultiFlowOrder(t *testing.T) {
	db := setupTestDB(t)
	flowRepo := repository.NewLinkFlowRepo(db)
	itemRepo := repository.NewLinkFlowItemRepo(db)
	linkRepo := repository.NewLinkRepo(db)
	svc := NewLinkFlowSvc(flowRepo, itemRepo, linkRepo)

	group := model.LinkGroup{Name: "G1"}
	db.Create(&group)
	flowA := model.LinkFlow{GroupID: group.ID, Name: "Morning"}
	flowB := model.LinkFlow{GroupID: group.ID, Name: "Release"}
	db.Create(&flowA)
	db.Create(&flowB)
	link := model.Link{Name: "GitHub", URL: "https://github.com", GroupID: &group.ID, OrderIndex: 20}
	db.Create(&link)

	orderA, orderB := 10, 30
	_, err := svc.AddLink(flowA.ID, model.LinkFlowLinkRequest{LinkID: link.ID.String(), OrderIndex: &orderA})
	require.NoError(t, err)
	_, err = svc.AddLink(flowB.ID, model.LinkFlowLinkRequest{LinkID: link.ID.String(), OrderIndex: &orderB})
	require.NoError(t, err)
	_, err = svc.UpdateLinkOrder(flowA.ID, link.ID, model.LinkFlowLinkOrderUpdate{OrderIndex: 40})
	require.NoError(t, err)

	itemA, err := itemRepo.Find(flowA.ID, link.ID)
	require.NoError(t, err)
	itemB, err := itemRepo.Find(flowB.ID, link.ID)
	require.NoError(t, err)
	storedLink, err := linkRepo.FindByID(link.ID)
	require.NoError(t, err)
	assert.Equal(t, 40, itemA.OrderIndex)
	assert.Equal(t, 30, itemB.OrderIndex)
	assert.Equal(t, 20, storedLink.OrderIndex)
}

func TestLinkFlowSvcRemoveLink(t *testing.T) {
	db := setupTestDB(t)
	flowRepo := repository.NewLinkFlowRepo(db)
	linkRepo := repository.NewLinkRepo(db)
	svc := NewLinkFlowSvc(flowRepo, repository.NewLinkFlowItemRepo(db), linkRepo)

	g := model.LinkGroup{Name: "G1"}
	db.Create(&g)
	flow := model.LinkFlow{GroupID: g.ID, Name: "F1"}
	db.Create(&flow)
	link := model.Link{Name: "L1", URL: "https://a.com", GroupID: &g.ID}
	db.Create(&link)
	itemRepo := repository.NewLinkFlowItemRepo(db)
	require.NoError(t, itemRepo.Add(&model.LinkFlowItem{FlowID: flow.ID, LinkID: link.ID}))

	// Removing membership preserves the link.
	err := svc.RemoveLink(flow.ID, link.ID)
	assert.NoError(t, err)
	var count int64
	db.Model(&model.Link{}).Where("id = ?", link.ID).Count(&count)
	assert.Equal(t, int64(1), count)
	_, err = itemRepo.Find(flow.ID, link.ID)
	assert.Error(t, err)

	// Remove link not in flow
	link2 := model.Link{Name: "L2", URL: "https://b.com", GroupID: &g.ID}
	db.Create(&link2)
	err = svc.RemoveLink(flow.ID, link2.ID)
	assert.Error(t, err)

	// Remove nonexistent link
	err = svc.RemoveLink(flow.ID, uuid.New())
	assert.Error(t, err)
}
