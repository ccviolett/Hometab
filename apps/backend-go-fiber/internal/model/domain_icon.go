package model

import "time"

type DomainIcon struct {
	Host               string    `gorm:"type:text;primaryKey" json:"host"`
	IconPath           string    `gorm:"type:text" json:"icon_path"`
	ContentType        string    `gorm:"type:text" json:"content_type"`
	Hash               string    `gorm:"type:text" json:"hash"`
	Source             string    `gorm:"type:text;not null;default:auto" json:"source"`
	Status             string    `gorm:"type:text;not null;default:failed" json:"status"`
	PendingPath        string    `gorm:"type:text" json:"pending_path"`
	PendingContentType string    `gorm:"type:text" json:"pending_content_type"`
	PendingHash        string    `gorm:"type:text" json:"pending_hash"`
	LastCheckedAt      time.Time `json:"last_checked_at"`
	ErrorMessage       string    `gorm:"type:text" json:"error_message"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type DomainIconCheckRequest struct {
	URL string `json:"url"`
}

type DomainIconCheckResponse struct {
	Host           string `json:"host"`
	Status         string `json:"status"`
	CurrentIconURL string `json:"current_icon_url,omitempty"`
	PendingIconURL string `json:"pending_icon_url,omitempty"`
	Error          string `json:"error,omitempty"`
}

type DomainIconChooseRequest struct {
	Choice string `json:"choice"`
}

type DomainIconRefreshAllResponse struct {
	TotalLinks int      `json:"total_links"`
	TotalHosts int      `json:"total_hosts"`
	Ready      int      `json:"ready"`
	Unchanged  int      `json:"unchanged"`
	Failed     int      `json:"failed"`
	Conflicts  int      `json:"conflicts"`
	Errors     []string `json:"errors,omitempty"`
}
