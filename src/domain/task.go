package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
}

func NewTask(t, d string) *Task {
	return &Task{
		Title:       t,
		Description: d,
	}
}

type Tasks struct {
	Data  []*Task
	Total int
}

func NewTasks(ts []*Task, total int) *Tasks {
	return &Tasks{
		Data:  ts,
		Total: total,
	}
}

type TaskUseCase interface {
	Fetch(ctx context.Context, f *Filter) (*Tasks, error)
	Insert(ctx context.Context, t *Task) error
	Update(ctx context.Context, uuid string, t *Task) error
	GetByID(ctx context.Context, uuid string) (*Task, error)
	Delete(ctx context.Context, uuid string) error
}

type TaskRepository interface {
	Fetch(ctx context.Context, f *Filter) (*Tasks, error)
	Insert(ctx context.Context, t *Task) error
	Update(ctx context.Context, uuid string, t *Task) error
	GetByID(ctx context.Context, uuid string) (*Task, error)
	Delete(ctx context.Context, uuid string) error
}
