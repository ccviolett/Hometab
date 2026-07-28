package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Link struct {
	ID         uuid.UUID  `gorm:"type:text;primaryKey" json:"id"`
	Name       string     `gorm:"not null" json:"name"`
	URL        string     `gorm:"not null" json:"url"`
	GroupID    *uuid.UUID `gorm:"type:text" json:"group_id,omitempty"`
	FlowID     *uuid.UUID `gorm:"type:text" json:"flow_id,omitempty"`
	OrderIndex int        `gorm:"default:0" json:"order_index"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (l *Link) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

type LinkCreate struct {
	Name       string  `json:"name" validate:"required"`
	URL        string  `json:"url" validate:"required"`
	GroupID    *string `json:"group_id,omitempty"`
	FlowID     *string `json:"flow_id,omitempty"`
	OrderIndex int     `json:"order_index"`
}

type LinkUpdate struct {
	Name       *string `json:"name,omitempty"`
	URL        *string `json:"url,omitempty"`
	GroupID    *string `json:"group_id,omitempty"`
	FlowID     *string `json:"flow_id,omitempty"`
	OrderIndex *int    `json:"order_index,omitempty"`
}
