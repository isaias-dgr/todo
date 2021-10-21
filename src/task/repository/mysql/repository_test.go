package mysql_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/isaias-dgr/todo/src/domain"
	"github.com/isaias-dgr/todo/src/task/repository/mysql"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

type SuiteRepository struct {
	suite.Suite
	db      *sql.DB
	mockSQL sqlmock.Sqlmock
	repo    domain.TaskRepository
}

func (s *SuiteRepository) SetupTest() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	db, mockSQL, err := sqlmock.New()
	if err != nil {
		s.Failf("an error '%s' was not expected when opening a stub database connection", err.Error())
	}
	s.db = db
	s.mockSQL = mockSQL
	s.repo = mysql.NewtaskRepository(db, sugar)
}

func (s *SuiteRepository) TestFetch() {

	s.Run("Success test", func(){
		mockTask := []*domain.Task{
			domain.NewTask("title 01", "description 01"),
			domain.NewTask("title 02", "description 02"),
		}

		binary_uuid, _ := uuid.New().MarshalBinary()
		rows := []string{"id", "title", "description", "updated_at", "created_at"}
		data := sqlmock.NewRows(rows).
			AddRow(binary_uuid, mockTask[0].Title, mockTask[0].Description,
				mockTask[0].CreatedAt, mockTask[0].UpdatedAt).
			AddRow(binary_uuid, mockTask[1].Title, mockTask[1].Description,
				mockTask[1].CreatedAt, mockTask[1].UpdatedAt)

		q := "SELECT \\* FROM task ORDER BY created_at ASC LIMIT \\? OFFSET \\?"
		s.mockSQL.ExpectQuery(q).WithArgs(3, 0).WillReturnRows(data)

		query_count := "SELECT count\\(\\*\\) FROM task"
		count := sqlmock.NewRows([]string{"count"}).AddRow(len(mockTask))
		s.mockSQL.ExpectQuery(query_count).WillReturnRows(count)

		filter := &domain.Filter{
			Offset: 0,
			Limit:  3,
			SortBy: "",
		}
		tasks, err := s.repo.Fetch(context.TODO(), filter)
		s.NoError(err)
		s.Equal(len(mockTask), tasks.Total)
		for i, task := range tasks.Data {
			s.Equal(task.Title, mockTask[i].Title)
			s.Equal(task.Description, mockTask[i].Description)
		}
	})

	s.Run("When exec query fails must return error", func(){
		q := "SELECT \\* FROM task ORDER BY created_at ASC LIMIT \\? OFFSET \\?"
		s.mockSQL.ExpectQuery(q).WithArgs(3, 0).WillReturnError(errors.New("D error"))
		filter := &domain.Filter{
			Offset: 0,
			Limit:  3,
			SortBy: "",
		}
		tasks, err := s.repo.Fetch(context.TODO(), filter)
		s.Error(err)
		s.Equal("query_context",err.Error())
		s.Nil(tasks)
	})

	s.Run("When db return incorrect type data", func(){
		rows := []string{"id", "title", "description", "updated_at", "created_at"}
		data := sqlmock.NewRows(rows).AddRow("uuid",  "T", "D", "C", "U")
		q := "SELECT \\* FROM task ORDER BY created_at ASC LIMIT \\? OFFSET \\?"
		s.mockSQL.ExpectQuery(q).WithArgs(3, 0).WillReturnRows(data)

		filter := &domain.Filter{
			Offset: 0,
			Limit:  3,
			SortBy: "",
		}
		tasks, err := s.repo.Fetch(context.TODO(), filter)
		s.Error(err)
		s.Equal("row_data_types",err.Error())
		s.Nil(tasks)
	})

	s.Run("When the row has a errors must return error", func() {
		mockTask := []*domain.Task{
			domain.NewTask("title 01", "description 01"),
			domain.NewTask("title 02", "description 02"),
		}

		binary_uuid, _ := uuid.New().MarshalBinary()
		rows := []string{"id", "title", "description", "updated_at", "created_at"}
		data := sqlmock.NewRows(rows).
			AddRow(binary_uuid, mockTask[0].Title, mockTask[0].Description,
				mockTask[0].CreatedAt, mockTask[0].UpdatedAt).
			AddRow(binary_uuid, mockTask[1].Title, mockTask[1].Description,
				mockTask[1].CreatedAt, mockTask[1].UpdatedAt).
			RowError(1, errors.New("row_error"))

		q := "SELECT \\* FROM task ORDER BY created_at ASC LIMIT \\? OFFSET \\?"
		s.mockSQL.ExpectQuery(q).WithArgs(3, 0).WillReturnRows(data)

		filter := &domain.Filter{
			Offset: 0,
			Limit:  3,
			SortBy: "",
		}
		_, err := s.repo.Fetch(context.TODO(), filter)
		s.Error(err)
		s.Equal("row_corrupt", err.Error())
	})

	s.Run("When Count exec query fails must return error", func(){
		mockTask := []*domain.Task{
			domain.NewTask("title 01", "description 01"),
			domain.NewTask("title 02", "description 02"),
		}

		binary_uuid, _ := uuid.New().MarshalBinary()
		rows := []string{"id", "title", "description", "updated_at", "created_at"}
		data := sqlmock.NewRows(rows).
			AddRow(binary_uuid, mockTask[0].Title, mockTask[0].Description,
				mockTask[0].CreatedAt, mockTask[0].UpdatedAt).
			AddRow(binary_uuid, mockTask[1].Title, mockTask[1].Description,
				mockTask[1].CreatedAt, mockTask[1].UpdatedAt)

		q := "SELECT \\* FROM task ORDER BY created_at ASC LIMIT \\? OFFSET \\?"
		s.mockSQL.ExpectQuery(q).WithArgs(3, 0).WillReturnRows(data)

		query_count := "SELECT count\\(\\*\\) FROM task"
		s.mockSQL.ExpectQuery(query_count).
			WillReturnError(errors.New("error_count"))

		filter := &domain.Filter{
			Offset: 0,
			Limit:  3,
			SortBy: "",
		}
		tasks, err := s.repo.Fetch(context.TODO(), filter)
		s.Error(err)
		s.Equal("query_context",err.Error())
		s.Nil(tasks)
	})

	s.Run("When Count exec query return incorrect type must return error", 
	func(){
		mockTask := []*domain.Task{
			domain.NewTask("title 01", "description 01"),
			domain.NewTask("title 02", "description 02"),
		}

		binary_uuid, _ := uuid.New().MarshalBinary()
		rows := []string{"id", "title", "description", "updated_at", "created_at"}
		data := sqlmock.NewRows(rows).
			AddRow(binary_uuid, mockTask[0].Title, mockTask[0].Description,
				mockTask[0].CreatedAt, mockTask[0].UpdatedAt).
			AddRow(binary_uuid, mockTask[1].Title, mockTask[1].Description,
				mockTask[1].CreatedAt, mockTask[1].UpdatedAt)

		q := "SELECT \\* FROM task ORDER BY created_at ASC LIMIT \\? OFFSET \\?"
		s.mockSQL.ExpectQuery(q).WithArgs(3, 0).WillReturnRows(data)

		query_count := "SELECT count\\(\\*\\) FROM task"
		count := sqlmock.NewRows([]string{"count"}).AddRow("a")
		s.mockSQL.ExpectQuery(query_count).WillReturnRows(count)

		filter := &domain.Filter{
			Offset: 0,
			Limit:  3,
			SortBy: "",
		}
		tasks, err := s.repo.Fetch(context.TODO(), filter)
		s.Error(err)
		s.Equal("row_data_types",err.Error())
		s.Nil(tasks)
	})

	s.Run("When Count has rows with error must return error", func(){
		mockTask := []*domain.Task{
			domain.NewTask("title 01", "description 01"),
			domain.NewTask("title 02", "description 02"),
		}

		binary_uuid, _ := uuid.New().MarshalBinary()
		rows := []string{"id", "title", "description", "updated_at", "created_at"}
		data := sqlmock.NewRows(rows).
			AddRow(binary_uuid, mockTask[0].Title, mockTask[0].Description,
				mockTask[0].CreatedAt, mockTask[0].UpdatedAt).
			AddRow(binary_uuid, mockTask[1].Title, mockTask[1].Description,
				mockTask[1].CreatedAt, mockTask[1].UpdatedAt)

		q := "SELECT \\* FROM task ORDER BY created_at ASC LIMIT \\? OFFSET \\?"
		s.mockSQL.ExpectQuery(q).WithArgs(3, 0).WillReturnRows(data)

		query_count := "SELECT count\\(\\*\\) FROM task"
		count := sqlmock.NewRows([]string{"count"}).
			AddRow(3).AddRow(4).
			RowError(1, errors.New("row_error"))
		s.mockSQL.ExpectQuery(query_count).WillReturnRows(count)

		filter := &domain.Filter{
			Offset: 0,
			Limit:  3,
			SortBy: "",
		}
		tasks, err := s.repo.Fetch(context.TODO(), filter)
		s.Error(err)
		s.Equal("row_corrupt", err.Error())
		s.Nil(tasks)
	})
}

func (s *SuiteRepository) TestGetByID() {
	s.Run("Success test return a task", func() {
		mockTask := domain.NewTask("title test 01", "description test 01")
		raw_uuid := uuid.New()
		binary_uuid, _ := raw_uuid.MarshalBinary()
		rows := []string{"id", "title", "description", "updated_at", "created_at"}
		data := sqlmock.NewRows(rows).
			AddRow(binary_uuid, mockTask.Title, mockTask.Description,
				mockTask.CreatedAt, mockTask.UpdatedAt)

		q := "SELECT \\* FROM task WHERE id=\\? "
		s.mockSQL.ExpectQuery(q).WithArgs(binary_uuid).WillReturnRows(data)

		task, err := s.repo.GetByID(context.TODO(), raw_uuid.String())
		s.NoError(err)
		s.Equal(mockTask.Title, task.Title)
		s.Equal(mockTask.Description, task.Description)
	})

	s.Run("When test uuid without format return error", func() {
		task, err := s.repo.GetByID(context.TODO(),"00000000-0000-0000-0000" )
		s.Error(err)
		s.Nil(task)
	})

	s.Run("When the query fails return error", func() {
		raw_uuid := uuid.New()
		binary_uuid, _ := raw_uuid.MarshalBinary()

		q := "SELECT \\* FROM task WHERE id=\\? "
		s.mockSQL.ExpectQuery(q).WithArgs(binary_uuid).WillReturnError(errors.New("generic error"))

		task, err := s.repo.GetByID(context.TODO(), raw_uuid.String())
		s.Error(err)
		s.Equal("query_context", err.Error())
		s.Nil(task)
	})

	s.Run("When the query not found task", func() {
		raw_uuid := uuid.New()
		binary_uuid, _ := raw_uuid.MarshalBinary()
		rows := []string{"title", "description", "updated_at", "created_at"}
		data := sqlmock.NewRows(rows)

		q := "SELECT \\* FROM task WHERE id=\\? "
		s.mockSQL.ExpectQuery(q).WithArgs(binary_uuid).WillReturnRows(data)

		task, err := s.repo.GetByID(context.TODO(), raw_uuid.String())
		s.Error(err)
		s.Equal("not_found",err.Error())
		s.Nil(task)
	})
}

func (s *SuiteRepository) TestInsert() {
	s.Run("Success test return a task", func() {
		task := domain.NewTask("Title new", "Description new")
		q := "INSERT task SET id=\\?, title=\\?, description=\\?, created_at=\\?, updated_at=\\?"
		s.mockSQL.
			ExpectPrepare(q).
			ExpectExec().
			WithArgs(sqlmock.AnyArg(), task.Title, task.Description,
				sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.repo.Insert(context.TODO(),task)
		s.Nil(err)
	})

	s.Run("When the prepare context faild must return error", func() {
		task := domain.NewTask("Title new", "Description new")
		q := "INSERT task SET id=\\?, title=\\?, description=\\?, created_at=\\?, updated_at=\\?"
		s.mockSQL.
			ExpectPrepare(q).
			WillReturnError(errors.New("prepare error"))

		err := s.repo.Insert(context.TODO(),task)
		s.Error(err)
		s.Equal("query_prepare_ctx", err.Error())
	})

	s.Run("When the Exec stmt faild must return error", func() {
		task := domain.NewTask("Title new", "Description new")

		q := "INSERT task SET id=\\?, title=\\?, description=\\?, created_at=\\?, updated_at=\\?"
		s.mockSQL.
			ExpectPrepare(q).
			ExpectExec().
			WithArgs(sqlmock.AnyArg(), task.Title, task.Description,
				sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("exec error"))

		err := s.repo.Insert(context.TODO(),task)
		s.Error(err)
		s.Equal("query_exec", err.Error())
	})

	s.Run("When the Exec result send error must return error", func() {
		task := domain.NewTask("title test 01", "description test 01")

		q := "INSERT task SET id=\\?, title=\\?, description=\\?, created_at=\\?, updated_at=\\?"
		s.mockSQL.ExpectPrepare(q).
			ExpectExec().
			WithArgs(sqlmock.AnyArg(), task.Title, task.Description, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("not_found")))

		err := s.repo.Insert(context.TODO(),task)
		s.NotNil(err)
		s.Equal("query_exec", err.Error())
	})

	s.Run("When the Exec insert more than one task must return error", func() {
		task := domain.NewTask("title test 01", "description test 01")

		q := "INSERT task SET id=\\?, title=\\?, description=\\?, created_at=\\?, updated_at=\\?"
		s.mockSQL.ExpectPrepare(q).
			ExpectExec().
			WithArgs(sqlmock.AnyArg(), task.Title, task.Description, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 2))
		err := s.repo.Insert(context.TODO(),task)
		s.NotNil(err)
		s.Equal("conflict_insert", err.Error())
	})
}

func (s *SuiteRepository) TestUpdate() {
	s.Run("Success test return a task", func() {
		task := domain.NewTask("title test 01", "description test 01")
		raw_uuid := uuid.New()
		binary_uuid, _ := raw_uuid.MarshalBinary()

		q := "UPDATE task set title=\\?, description=\\?, updated_at=\\? WHERE ID = \\?"
		s.mockSQL.
			ExpectPrepare(q).
			ExpectExec().
			WithArgs(task.Title, task.Description, sqlmock.AnyArg(), binary_uuid).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.repo.Update(context.TODO(), raw_uuid.String(), task)
		s.Nil(err)
	})

	s.Run("When test uuid without format return error", func() {
		mocktask := domain.NewTask("title test 01", "description test 01")
		err := s.repo.Update(context.TODO(),"00000000", mocktask)
		s.Error(err)
		s.Equal("uuid_format", err.Error())
	})

	s.Run("When the prepare context faild must return error", func() {
		task := domain.NewTask("title test 01", "description test 01")
		raw_uuid := uuid.New()

		q := "UPDATE task set title=\\?, description=\\?, updated_at=\\? WHERE ID = \\?"
		s.mockSQL.ExpectPrepare(q).WillReturnError(errors.New("prepare error"))
		err := s.repo.Update(context.TODO(), raw_uuid.String(), task)
		s.NotNil(err)
		s.Equal("query_prepare_ctx", err.Error())
	})

	s.Run("When the Exec stmt faild must return error", func() {
		task := domain.NewTask("title test 01", "description test 01")
		raw_uuid := uuid.New()
		binary_uuid, _ := raw_uuid.MarshalBinary()

		q := "UPDATE task set title=\\?, description=\\?, updated_at=\\? WHERE ID = \\?"
		s.mockSQL.ExpectPrepare(q).
			ExpectExec().
			WithArgs(task.Title, task.Description, sqlmock.AnyArg(), binary_uuid).
			WillReturnError(errors.New("exec error"))
		err := s.repo.Update(context.TODO(), raw_uuid.String(), task)
		s.NotNil(err)
		s.Equal("query_exec", err.Error())
	})

	s.Run("When the Exec result send error must return error", func() {
		task := domain.NewTask("title test 01", "description test 01")
		raw_uuid := uuid.New()
		binary_uuid, _ := raw_uuid.MarshalBinary()

		q := "UPDATE task set title=\\?, description=\\?, updated_at=\\? WHERE ID = \\?"
		s.mockSQL.ExpectPrepare(q).
			ExpectExec().
			WithArgs(task.Title, task.Description, sqlmock.AnyArg(), binary_uuid).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("not_found")))
		err := s.repo.Update(context.TODO(), raw_uuid.String(), task)
		s.NotNil(err)
		s.Equal("not_found", err.Error())
	})

	s.Run("When the Exec update more than one task must return error", func() {
		task := domain.NewTask("title test 01", "description test 01")
		raw_uuid := uuid.New()
		binary_uuid, _ := raw_uuid.MarshalBinary()

		q := "UPDATE task set title=\\?, description=\\?, updated_at=\\? WHERE ID = \\?"
		s.mockSQL.ExpectPrepare(q).
			ExpectExec().
			WithArgs(task.Title, task.Description, sqlmock.AnyArg(), binary_uuid).
			WillReturnResult(sqlmock.NewResult(1, 2))
		err := s.repo.Update(context.TODO(), raw_uuid.String(), task)
		s.NotNil(err)
		s.Equal("conflict_update", err.Error())
	})
}

func (s *SuiteRepository) TestDelete() {
	s.Run("Success test return a task", func() {
		raw_uuid := uuid.New()
		binary_uuid, _ := raw_uuid.MarshalBinary()
		q := "DELETE FROM task WHERE id=\\?"
		s.mockSQL.
			ExpectPrepare(q).
			ExpectExec().
			WithArgs(binary_uuid).
			WillReturnResult(sqlmock.NewResult(1, 1))
		err := s.repo.Delete(context.TODO(), raw_uuid.String())
		s.Nil(err)
	})

	s.Run("When test uuid without format return error", func() {
		err := s.repo.Delete(context.TODO(),"00000000")
		s.Error(err)
		s.Equal("uuid_format", err.Error())
	})

	s.Run("When the prepare context faild must return error", func() {
		raw_uuid := uuid.New()
		q := "DELETE FROM task WHERE id=\\?"
		s.mockSQL.ExpectPrepare(q).WillReturnError(errors.New("prepare error"))
		err := s.repo.Delete(context.TODO(), raw_uuid.String())
		s.Error(err)
		s.Equal("query_prepare_ctx", err.Error())
	})

	s.Run("When the Exec stmt faild must return error", func() {
		raw_uuid := uuid.New()
		binary_uuid, _ := raw_uuid.MarshalBinary()

		q := "DELETE FROM task WHERE id=\\?"
		s.mockSQL.ExpectPrepare(q).
			ExpectExec().
			WithArgs(binary_uuid).
			WillReturnError(errors.New("exec error"))
		err := s.repo.Delete(context.TODO(), raw_uuid.String())
		s.NotNil(err)
		s.Equal("query_exec", err.Error())
	})

	s.Run("When the Exec result send error must return error", func() {
		raw_uuid := uuid.New()
		binary_uuid, _ := raw_uuid.MarshalBinary()

		q := "DELETE FROM task WHERE id=\\?"
		s.mockSQL.ExpectPrepare(q).
			ExpectExec().
			WithArgs(binary_uuid).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("not_found")))
		err := s.repo.Delete(context.TODO(), raw_uuid.String())
		s.NotNil(err)
		s.Equal("query_exec_delete", err.Error())
	})

	s.Run("When the Exec update more than one task must return error", func() {
		raw_uuid := uuid.New()
		binary_uuid, _ := raw_uuid.MarshalBinary()

		q := "DELETE FROM task WHERE id=\\?"
		s.mockSQL.ExpectPrepare(q).
			ExpectExec().
			WithArgs(binary_uuid).
			WillReturnResult(sqlmock.NewResult(1, 2))
		err := s.repo.Delete(context.TODO(), raw_uuid.String())
		s.NotNil(err)
		s.Equal("conflict_delete", err.Error())
	})
}

func TestSuiteRepository(t *testing.T) {
	suite.Run(t, new(SuiteRepository))
}
