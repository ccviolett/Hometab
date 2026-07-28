package model

import (
	"time"
)

type SearchEngine struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string     `gorm:"not null" json:"name"`
	URLTemplate string     `gorm:"not null;column:url_template" json:"url_template"`
	Icon        *string    `json:"icon,omitempty"`
	Description *string    `json:"description,omitempty"`
	Color       *string    `json:"color,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type SearchEngineCreate struct {
	Name        string  `json:"name" validate:"required"`
	URLTemplate string  `json:"url_template" validate:"required"`
	Icon        *string `json:"icon,omitempty"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
}

type SearchEngineUpdate struct {
	Name        *string `json:"name,omitempty"`
	URLTemplate *string `json:"url_template,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
}
