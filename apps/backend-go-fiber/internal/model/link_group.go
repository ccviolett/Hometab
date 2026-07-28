package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LinkGroup struct {
	ID          uuid.UUID `gorm:"type:text;primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Description *string   `json:"description,omitempty"`
	OrderIndex  int       `gorm:"default:0" json:"order_index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (g *LinkGroup) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
}

type LinkGroupCreate struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description,omitempty"`
	OrderIndex  int     `json:"order_index"`
}

type LinkGroupUpdate struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	OrderIndex  *int    `json:"order_index,omitempty"`
}
