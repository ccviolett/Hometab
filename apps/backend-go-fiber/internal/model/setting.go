package model

import (
	"encoding/json"
	"time"
)

type Setting struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Key       string    `gorm:"uniqueIndex;not null" json:"key"`
	ValueJSON string    `gorm:"type:text;not null;column:value_json" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SettingCreate struct {
	Key   string      `json:"key" validate:"required"`
	Value interface{} `json:"value" validate:"required"`
}

type SettingUpdate struct {
	Value interface{} `json:"value" validate:"required"`
}

type SettingRead struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func (s *Setting) GetValue() interface{} {
	var v interface{}
	_ = json.Unmarshal([]byte(s.ValueJSON), &v)
	return v
}

func ValueToJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
