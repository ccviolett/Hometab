package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestLinkBeforeCreate(t *testing.T) {
	l := &Link{}
	err := l.BeforeCreate(&gorm.DB{})
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, l.ID)
}

func TestLinkGroupBeforeCreate(t *testing.T) {
	g := &LinkGroup{}
	err := g.BeforeCreate(&gorm.DB{})
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, g.ID)
}

func TestLinkFlowBeforeCreate(t *testing.T) {
	f := &LinkFlow{}
	err := f.BeforeCreate(&gorm.DB{})
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, f.ID)
}

func TestSettingGetValue(t *testing.T) {
	s := &Setting{ValueJSON: `"hello"`}
	assert.Equal(t, "hello", s.GetValue())

	s2 := &Setting{ValueJSON: `42`}
	assert.Equal(t, float64(42), s2.GetValue())

	s3 := &Setting{ValueJSON: `{"key":"val"}`}
	v := s3.GetValue()
	m, ok := v.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "val", m["key"])
}

func TestValueToJSON(t *testing.T) {
	raw := ValueToJSON("test")
	assert.Equal(t, `"test"`, raw)

	raw2 := ValueToJSON(42)
	assert.Equal(t, `42`, raw2)
}
