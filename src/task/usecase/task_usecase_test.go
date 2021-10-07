package useCase_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/isaias-dgr/todo/src/domain"
	"github.com/isaias-dgr/todo/src/domain/mocks"
	useCase "github.com/isaias-dgr/todo/src/task/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UseCaseSuite struct {
	suite.Suite
	repo *mocks.TaskRepository
	cu   domain.TaskUseCase
}

func (s *UseCaseSuite) SetupTest() {
	s.repo = new(mocks.TaskRepository)
	s.cu = useCase.NewTaskUseCase(s.repo)
}

func (s *UseCaseSuite) TestFetch() {
	s.repo.On("Fetch", mock.Anything, mock.Anything).Return(nil, nil)
	ctx := context.Background()
	val := url.Values{
		"offset":  []string{"10"},
		"limit":   []string{"10"},
		"sort_by": []string{""},
	}
	filter := domain.NewFilter(val)
	_, err := s.cu.Fetch(ctx, filter)
	assert.Nil(s.T(), err, "The fetch mock its not working")
}

func (s *UseCaseSuite) TestGetByID() {
	s.repo.On("GetByID", mock.Anything, mock.Anything).Return(nil, nil)
	ctx := context.Background()
	_, err := s.cu.GetByID(ctx, "000-0000")
	assert.Nil(s.T(), err, "The get mock its not working")
}

func (s *UseCaseSuite) TestUpdate() {
	s.repo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	ctx := context.Background()
	task := domain.Task{}
	err := s.cu.Update(ctx, "000-0000", &task)
	assert.Nil(s.T(), err, "The get mock its not working")
}

func (s *UseCaseSuite) TestInsert() {
	s.repo.On("Insert", mock.Anything, mock.Anything).Return(nil)
	ctx := context.Background()
	task := domain.Task{}
	err := s.cu.Insert(ctx, &task)
	assert.Nil(s.T(), err, "The get mock its not working")
}

func (s *UseCaseSuite) TestDelete() {
	s.repo.On("Delete", mock.Anything, mock.Anything).Return(nil, nil)
	ctx := context.Background()
	err := s.cu.Delete(ctx, "000-0000")
	assert.Nil(s.T(), err, "The get mock its not working")
}

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseSuite))
}
