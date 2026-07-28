package model

import (
	"time"

	"github.com/google/uuid"
)

// LinkFlowItem stores flow membership separately from a link's group order.
type LinkFlowItem struct {
	FlowID     uuid.UUID `gorm:"type:text;primaryKey;index" json:"flow_id"`
	LinkID     uuid.UUID `gorm:"type:text;primaryKey;index" json:"link_id"`
	OrderIndex int       `gorm:"not null;default:0" json:"order_index"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
