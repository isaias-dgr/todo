package useCase

import (
	"context"

	"github.com/isaias-dgr/todo/src/domain"
)

type taskUseCase struct {
	repo domain.TaskRepository
}

func NewTaskUseCase(t domain.TaskRepository) domain.TaskUseCase {
	return &taskUseCase{
		repo: t,
	}
}

func (t *taskUseCase) Fetch(ctx context.Context, f *domain.Filter) (ts *domain.Tasks, err error) {
	return t.repo.Fetch(ctx, f)
}

func (t *taskUseCase) GetByID(ctx context.Context, uuid string) (ta *domain.Task, err error) {
	return t.repo.GetByID(ctx, uuid)
}

func (t *taskUseCase) Update(ctx context.Context, uuid string, ta *domain.Task) (err error) {
	return t.repo.Update(ctx, uuid, ta)
}

func (t *taskUseCase) Insert(ctx context.Context, ta *domain.Task) (err error) {
	return t.repo.Insert(ctx, ta)
}

func (t *taskUseCase) Delete(ctx context.Context, uuid string) (err error) {
	return t.repo.Delete(ctx, uuid)
}
