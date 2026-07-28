package service

import (
	"testing"

	"hometab/internal/model"
	"hometab/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkSvcCreateWithGroup(t *testing.T) {
	db := setupTestDB(t)
	linkRepo := repository.NewLinkRepo(db)
	flowRepo := repository.NewLinkFlowRepo(db)
	groupRepo := repository.NewLinkGroupRepo(db)
	svc := NewLinkSvc(linkRepo, flowRepo, repository.NewLinkFlowItemRepo(db))

	// Create group
	g := model.LinkGroup{Name: "G1"}
	db.Create(&g)
	gid := g.ID.String()

	// Create link with group
	link, err := svc.Create(model.LinkCreate{
		Name: "L1", URL: "https://a.com", GroupID: &gid,
	})
	require.NoError(t, err)
	assert.Equal(t, g.ID, *link.GroupID)

	// Create with invalid group_id
	bad := "not-a-uuid"
	_, err = svc.Create(model.LinkCreate{Name: "L2", URL: "https://b.com", GroupID: &bad})
	assert.Error(t, err)

	// Create with flow_id
	flow := model.LinkFlow{GroupID: g.ID, Name: "F1"}
	db.Create(&flow)
	fid := flow.ID.String()
	link2, err := svc.Create(model.LinkCreate{
		Name: "L3", URL: "https://c.com", FlowID: &fid,
	})
	require.NoError(t, err)
	assert.Nil(t, link2.FlowID)
	assert.Equal(t, g.ID, *link2.GroupID)
	_, err = repository.NewLinkFlowItemRepo(db).Find(flow.ID, link2.ID)
	require.NoError(t, err)

	// Create with flow_id and matching group_id
	link3, err := svc.Create(model.LinkCreate{
		Name: "L4", URL: "https://d.com", FlowID: &fid, GroupID: &gid,
	})
	require.NoError(t, err)
	assert.Nil(t, link3.FlowID)
	_, err = repository.NewLinkFlowItemRepo(db).Find(flow.ID, link3.ID)
	require.NoError(t, err)

	// Create with flow_id and mismatched group_id
	g2 := model.LinkGroup{Name: "G2"}
	db.Create(&g2)
	gid2 := g2.ID.String()
	_, err = svc.Create(model.LinkCreate{
		Name: "L5", URL: "https://e.com", FlowID: &fid, GroupID: &gid2,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not match")

	// Create with invalid flow_id
	badFlow := "not-uuid"
	_, err = svc.Create(model.LinkCreate{Name: "L6", URL: "https://f.com", FlowID: &badFlow})
	assert.Error(t, err)

	// Create with nonexistent flow_id
	nonexist := uuid.New().String()
	_, err = svc.Create(model.LinkCreate{Name: "L7", URL: "https://g.com", FlowID: &nonexist})
	assert.Error(t, err)

	// FindAll
	items, err := svc.FindAll()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(items), 3)

	// FindByID
	got, err := svc.FindByID(link.ID)
	require.NoError(t, err)
	assert.Equal(t, "L1", got.Name)

	// FindByID not found
	_, err = svc.FindByID(uuid.New())
	assert.Error(t, err)

	_ = groupRepo // used for setup
}

func TestLinkSvcUpdate(t *testing.T) {
	db := setupTestDB(t)
	linkRepo := repository.NewLinkRepo(db)
	flowRepo := repository.NewLinkFlowRepo(db)
	svc := NewLinkSvc(linkRepo, flowRepo, repository.NewLinkFlowItemRepo(db))

	g := model.LinkGroup{Name: "G1"}
	db.Create(&g)
	gid := g.ID.String()

	link := model.Link{Name: "L1", URL: "https://a.com", GroupID: &g.ID}
	db.Create(&link)

	// Update name
	newName := "Updated"
	updated, err := svc.Update(link.ID, model.LinkUpdate{Name: &newName})
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Name)

	// Update URL
	newURL := "https://b.com"
	updated, err = svc.Update(link.ID, model.LinkUpdate{URL: &newURL})
	require.NoError(t, err)
	assert.Equal(t, "https://b.com", updated.URL)

	// Update order_index
	idx := 5
	updated, err = svc.Update(link.ID, model.LinkUpdate{OrderIndex: &idx})
	require.NoError(t, err)
	assert.Equal(t, 5, updated.OrderIndex)

	// Update group_id
	g2 := model.LinkGroup{Name: "G2"}
	db.Create(&g2)
	gid2 := g2.ID.String()
	updated, err = svc.Update(link.ID, model.LinkUpdate{GroupID: &gid2})
	require.NoError(t, err)
	assert.Equal(t, g2.ID, *updated.GroupID)

	// Update with invalid group_id
	badGid := "not-uuid"
	_, err = svc.Update(link.ID, model.LinkUpdate{GroupID: &badGid})
	assert.Error(t, err)

	// Update flow_id to valid flow
	flow := model.LinkFlow{GroupID: g2.ID, Name: "F1"}
	db.Create(&flow)
	fid := flow.ID.String()
	updated, err = svc.Update(link.ID, model.LinkUpdate{FlowID: &fid})
	require.NoError(t, err)
	assert.Nil(t, updated.FlowID)
	_, err = repository.NewLinkFlowItemRepo(db).Find(flow.ID, link.ID)
	require.NoError(t, err)

	// Empty legacy flow_id no longer removes memberships; use the flow membership API.
	empty := ""
	updated, err = svc.Update(link.ID, model.LinkUpdate{FlowID: &empty})
	require.NoError(t, err)
	assert.Nil(t, updated.FlowID)

	// Update with invalid flow_id
	badFid := "not-uuid"
	_, err = svc.Update(link.ID, model.LinkUpdate{FlowID: &badFid})
	assert.Error(t, err)

	// Update with nonexistent flow_id
	nonexist := uuid.New().String()
	_, err = svc.Update(link.ID, model.LinkUpdate{FlowID: &nonexist})
	assert.Error(t, err)

	// Update with mismatched group/flow
	flow2 := model.LinkFlow{GroupID: g.ID, Name: "F2"}
	db.Create(&flow2)
	fid2 := flow2.ID.String()
	// link is now in g2, flow2 is in g (mismatch)
	_, err = svc.Update(link.ID, model.LinkUpdate{FlowID: &fid2})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not match")

	// Update not found
	_, err = svc.Update(uuid.New(), model.LinkUpdate{Name: &newName})
	assert.Error(t, err)

	// Delete
	err = svc.Delete(link.ID)
	assert.NoError(t, err)
	_ = gid
}
