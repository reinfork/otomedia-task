package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	StatusTodo       = "todo"
	StatusInProgress = "in_progress"
	StatusDone       = "done"
)

// Task is the core domain model.
// Soft delete is implemented via gorm.DeletedAt (deleted_at IS NULL = visible).
type Task struct {
	ID          uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	Title       string         `gorm:"size:255;not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Status      string         `gorm:"size:32;not null;default:todo;index" json:"status"`
	Assignee    string         `gorm:"size:255;index" json:"assignee"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (t *Task) BeforeCreate(_ *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.Status == "" {
		t.Status = StatusTodo
	}
	return nil
}

func IsValidStatus(s string) bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

// TaskFilter carries list-query parameters.
type TaskFilter struct {
	Status   string
	Keyword  string
	Assignee string
	Page     int
	Limit    int
	Sort     string
}

func (f *TaskFilter) Normalize() {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 10
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	if f.Sort == "" {
		f.Sort = "-created_at"
	}
}

func (f TaskFilter) Offset() int {
	return (f.Page - 1) * f.Limit
}
