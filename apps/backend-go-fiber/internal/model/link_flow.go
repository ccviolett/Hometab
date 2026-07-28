package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LinkFlow struct {
	ID         uuid.UUID `gorm:"type:text;primaryKey" json:"id"`
	GroupID    uuid.UUID `gorm:"type:text;not null" json:"group_id"`
	Name       string    `gorm:"not null" json:"name"`
	OrderIndex int       `gorm:"default:0" json:"order_index"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (f *LinkFlow) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

type LinkFlowCreate struct {
	GroupID    string `json:"group_id" validate:"required"`
	Name       string `json:"name" validate:"required"`
	OrderIndex *int   `json:"order_index,omitempty"`
}

type LinkFlowUpdate struct {
	GroupID    *string `json:"group_id,omitempty"`
	Name       *string `json:"name,omitempty"`
	OrderIndex *int    `json:"order_index,omitempty"`
}

type LinkFlowDeleteOptions struct {
	LinkIDsToKeep []string `json:"link_ids_to_keep"`
}

type LinkFlowLinkRequest struct {
	LinkID     string `json:"link_id" validate:"required"`
	OrderIndex *int   `json:"order_index,omitempty"`
}

type LinkFlowLinkOrderUpdate struct {
	OrderIndex int `json:"order_index"`
}

type ReorderRequest struct {
	IDs []string `json:"ids"`
}

type LinkFlowWithLinks struct {
	Flow  LinkFlow `json:"flow"`
	Links []Link   `json:"links"`
}

type GroupedLinksResponse struct {
	Group LinkGroup           `json:"group"`
	Flows []LinkFlowWithLinks `json:"flows"`
	Links []Link              `json:"links"`
}
