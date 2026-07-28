package repository

import (
	"errors"

	"hometab/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LinkFlowItemRepo struct {
	db *gorm.DB
}

func NewLinkFlowItemRepo(db *gorm.DB) *LinkFlowItemRepo {
	return &LinkFlowItemRepo{db: db}
}

func (r *LinkFlowItemRepo) Add(item *model.LinkFlowItem) error {
	var existing model.LinkFlowItem
	err := r.db.First(&existing, "flow_id = ? AND link_id = ?", item.FlowID, item.LinkID).Error
	if err == nil {
		return errors.New("link already belongs to this flow")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return r.db.Create(item).Error
}

func (r *LinkFlowItemRepo) Find(flowID, linkID uuid.UUID) (*model.LinkFlowItem, error) {
	var item model.LinkFlowItem
	if err := r.db.First(&item, "flow_id = ? AND link_id = ?", flowID, linkID).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *LinkFlowItemRepo) FindLinks(flowID uuid.UUID) ([]model.Link, error) {
	items := make([]model.Link, 0)
	err := r.db.Table("links").
		Select("links.*").
		Joins("JOIN link_flow_items ON link_flow_items.link_id = links.id").
		Where("link_flow_items.flow_id = ?", flowID).
		Order("link_flow_items.order_index, link_flow_items.created_at, links.created_at").
		Find(&items).Error
	return items, err
}

func (r *LinkFlowItemRepo) UpdateOrder(flowID, linkID uuid.UUID, orderIndex int) error {
	result := r.db.Model(&model.LinkFlowItem{}).
		Where("flow_id = ? AND link_id = ?", flowID, linkID).
		Update("order_index", orderIndex)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *LinkFlowItemRepo) Remove(flowID, linkID uuid.UUID) error {
	result := r.db.Delete(&model.LinkFlowItem{}, "flow_id = ? AND link_id = ?", flowID, linkID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *LinkFlowItemRepo) DeleteByFlowID(flowID uuid.UUID) error {
	return r.db.Delete(&model.LinkFlowItem{}, "flow_id = ?", flowID).Error
}

func (r *LinkFlowItemRepo) DeleteByLinkID(linkID uuid.UUID) error {
	return r.db.Delete(&model.LinkFlowItem{}, "link_id = ?", linkID).Error
}

func (r *LinkFlowItemRepo) Reorder(flowID uuid.UUID, ids []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for index, id := range ids {
			result := tx.Model(&model.LinkFlowItem{}).
				Where("flow_id = ? AND link_id = ?", flowID, id).
				Update("order_index", index*10)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return gorm.ErrRecordNotFound
			}
		}
		return nil
	})
}
