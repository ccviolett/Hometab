package repository

import (
	"hometab/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LinkRepo struct {
	db *gorm.DB
}

func NewLinkRepo(db *gorm.DB) *LinkRepo {
	return &LinkRepo{db: db}
}

func (r *LinkRepo) FindAll() ([]model.Link, error) {
	items := make([]model.Link, 0)
	err := r.db.Order("order_index, created_at").Find(&items).Error
	return items, err
}

func (r *LinkRepo) FindByID(id uuid.UUID) (*model.Link, error) {
	var item model.Link
	err := r.db.First(&item, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *LinkRepo) FindByGroupID(gid uuid.UUID) ([]model.Link, error) {
	items := make([]model.Link, 0)
	err := r.db.Where("group_id = ?", gid).Order("order_index, created_at").Find(&items).Error
	return items, err
}

func (r *LinkRepo) FindUngroupedByGroupID(gid uuid.UUID) ([]model.Link, error) {
	items := make([]model.Link, 0)
	err := r.db.Where("group_id = ? AND (flow_id IS NULL)", gid).Order("order_index, created_at").Find(&items).Error
	return items, err
}

func (r *LinkRepo) FindByFlowID(fid uuid.UUID) ([]model.Link, error) {
	items := make([]model.Link, 0)
	err := r.db.Where("flow_id = ?", fid).Order("order_index, created_at").Find(&items).Error
	return items, err
}

func (r *LinkRepo) Create(item *model.Link) error {
	return r.db.Create(item).Error
}

func (r *LinkRepo) Update(item *model.Link) error {
	return r.db.Save(item).Error
}

func (r *LinkRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Link{}, "id = ?", id).Error
}

func (r *LinkRepo) MoveToGroup(fromGroupID, toGroupID uuid.UUID) error {
	return r.db.Model(&model.Link{}).Where("group_id = ?", fromGroupID).Updates(map[string]interface{}{
		"group_id": toGroupID,
		"flow_id":  nil,
	}).Error
}

func (r *LinkRepo) ClearFlowID(flowID uuid.UUID, keepIDs []uuid.UUID) error {
	q := r.db.Model(&model.Link{}).Where("flow_id = ?", flowID)
	if len(keepIDs) > 0 {
		q = q.Where("id IN ?", keepIDs)
	}
	return q.Update("flow_id", nil).Error
}

func (r *LinkRepo) DeleteByFlowID(flowID uuid.UUID, excludeIDs []uuid.UUID) error {
	q := r.db.Where("flow_id = ?", flowID)
	if len(excludeIDs) > 0 {
		q = q.Where("id NOT IN ?", excludeIDs)
	}
	return q.Delete(&model.Link{}).Error
}

func (r *LinkRepo) ReorderGroup(groupID uuid.UUID, ids []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for index, id := range ids {
			result := tx.Model(&model.Link{}).Where("id = ? AND group_id = ?", id, groupID).Update("order_index", index*10)
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
