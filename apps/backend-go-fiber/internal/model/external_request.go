package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ExternalRequest struct {
	ID               uuid.UUID `gorm:"type:text;primaryKey" json:"id"`
	Name             string    `gorm:"not null" json:"name"`
	Description      string    `json:"description"`
	Method           string    `gorm:"not null;default:GET" json:"method"`
	URL              string    `gorm:"not null" json:"url"`
	HeadersJSON      string    `gorm:"type:text" json:"headers_json"`
	QueryJSON        string    `gorm:"type:text" json:"query_json"`
	BodyType         string    `gorm:"not null;default:none" json:"body_type"`
	Body             string    `gorm:"type:text" json:"body"`
	ParserType       string    `gorm:"not null;default:status" json:"parser_type"`
	ParserConfigJSON string    `gorm:"type:text" json:"parser_config_json"`
	ConfirmBeforeRun bool      `gorm:"default:false" json:"confirm_before_run"`
	Enabled          bool      `gorm:"default:true" json:"enabled"`
	OrderIndex       int       `gorm:"default:0" json:"order_index"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (r *ExternalRequest) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

type ExternalRequestCreate struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	Method           string `json:"method"`
	URL              string `json:"url"`
	HeadersJSON      string `json:"headers_json"`
	QueryJSON        string `json:"query_json"`
	BodyType         string `json:"body_type"`
	Body             string `json:"body"`
	ParserType       string `json:"parser_type"`
	ParserConfigJSON string `json:"parser_config_json"`
	ConfirmBeforeRun *bool  `json:"confirm_before_run"`
	Enabled          *bool  `json:"enabled"`
	OrderIndex       int    `json:"order_index"`
}

type ExternalRequestUpdate struct {
	Name             *string `json:"name,omitempty"`
	Description      *string `json:"description,omitempty"`
	Method           *string `json:"method,omitempty"`
	URL              *string `json:"url,omitempty"`
	HeadersJSON      *string `json:"headers_json,omitempty"`
	QueryJSON        *string `json:"query_json,omitempty"`
	BodyType         *string `json:"body_type,omitempty"`
	Body             *string `json:"body,omitempty"`
	ParserType       *string `json:"parser_type,omitempty"`
	ParserConfigJSON *string `json:"parser_config_json,omitempty"`
	ConfirmBeforeRun *bool   `json:"confirm_before_run,omitempty"`
	Enabled          *bool   `json:"enabled,omitempty"`
	OrderIndex       *int    `json:"order_index,omitempty"`
}

type ExternalRequestParsedField struct {
	Label string      `json:"label"`
	Path  string      `json:"path,omitempty"`
	Value interface{} `json:"value,omitempty"`
	Error string      `json:"error,omitempty"`
}

type ExternalRequestExecuteResult struct {
	Status      int                          `json:"status"`
	StatusText  string                       `json:"status_text"`
	DurationMS  int64                        `json:"duration_ms"`
	Headers     map[string][]string          `json:"headers"`
	BodyPreview string                       `json:"body_preview"`
	Parsed      []ExternalRequestParsedField `json:"parsed"`
	Error       string                       `json:"error,omitempty"`
}
