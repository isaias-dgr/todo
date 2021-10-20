package http_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/isaias-dgr/todo/src/domain"
	"github.com/isaias-dgr/todo/src/domain/mocks"
	h "github.com/isaias-dgr/todo/src/task/deliver/http"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type SuiteTodo struct {
	suite.Suite
	cu      *mocks.TaskUseCase
	handler *h.TaskHandler
}

func (s *SuiteTodo) SetupTest() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()
	s.cu = new(mocks.TaskUseCase)
	s.handler = &h.TaskHandler{
		TuseCase: s.cu,
		L:        sugar,
	}
}

func (s *SuiteTodo) TestFetch() {
	s.Run("When the use case is succesful", func() {
		tasks := domain.NewTasks(
			[]*domain.Task{
				domain.NewTask("title 1", "description 1"),
				domain.NewTask("title 2", "description 2"),
				domain.NewTask("title 3", "description 3"),
			},
			3,
		)
		filter := domain.Filter{
			Offset: 1,
			Limit:  3,
			SortBy: "",
		}
		s.cu.On("Fetch", mock.Anything, &filter).Return(tasks, nil)
		req, err := http.NewRequest("GET", "/task/?offset=1&limit=3", strings.NewReader(""))
		s.NoError(err)
		w := httptest.NewRecorder()
		s.handler.FetchTasks(w, req)
		s.Equal(http.StatusOK, w.Code)
		var response domain.Response
		err = json.Unmarshal(w.Body.Bytes(), &response)
		s.NoError(err)
		expected := "{\"data\":[{\"id\":\"00000000-0000-0000-0000-000000000000\",\"title\":\"title 1\",\"description\":\"description 1\"},{\"id\":\"00000000-0000-0000-0000-000000000000\",\"title\":\"title 2\",\"description\":\"description 2\"},{\"id\":\"00000000-0000-0000-0000-000000000000\",\"title\":\"title 3\",\"description\":\"description 3\"}],\"metadata\":{\"offset\":1,\"limit\":3,\"total\":3}}"
		s.Equal(expected, w.Body.String())
	})

	s.Run("When the use case return a generic error", func() {
		s.cu.On("Fetch", mock.Anything, mock.Anything).
			Return(nil, errors.New("error"))
		req, err := http.NewRequest("GET", "/task/", strings.NewReader(""))
		s.NoError(err)
		w := httptest.NewRecorder()

		s.handler.FetchTasks(w, req)
		s.Equal(http.StatusInternalServerError, w.Code)
		msg := []byte("{\"message\":\"error\"}")
		s.Equal(msg, w.Body.Bytes())
	})
}

func (s *SuiteTodo) TestInsert() {
	s.Run("When the use case is succesful", func() {
		task := domain.NewTask("t001", "td00001")
		s.cu.On("Insert", mock.Anything, task).Return(nil)
		req, err := http.NewRequest("POST", "/task/", strings.NewReader("{\"title\": \"t001\",\"description\": \"td00001\"}"))
		s.NoError(err)
		w := httptest.NewRecorder()
		s.handler.InsertTask(w, req)
		s.Equal(http.StatusAccepted, w.Code)
		s.NoError(err)
		expected := "{\"data\":{\"id\":\"00000000-0000-0000-0000-000000000000\",\"title\":\"t001\",\"description\":\"td00001\"}}"
		s.Equal(expected, w.Body.String())
	})

	s.Run("When the use case return a generic error", func() {
		task := domain.NewTask("t002", "td00002")
		s.cu.On("Insert", mock.Anything, task).Return(errors.New("G Error"))
		req, err := http.NewRequest("POST", "/task/", strings.NewReader("{\"title\": \"t002\",\"description\": \"td00002\"}"))
		s.NoError(err)
		w := httptest.NewRecorder()
		s.handler.InsertTask(w, req)
		s.Equal(http.StatusBadRequest, w.Code)
		s.NoError(err)
		expected := "{\"message\":\"G Error\"}"
		s.Equal(expected, w.Body.String())
	})

	s.Run("When the payload has a error", func() {
		s.cu.On("Insert", mock.Anything, mock.Anything).Return(nil)
		req, err := http.NewRequest("POST", "/task/", strings.NewReader("{\"title\": \"t002\",\"description\": \"td00002}"))
		s.NoError(err)
		w := httptest.NewRecorder()
		s.handler.InsertTask(w, req)
		s.Equal(http.StatusBadRequest, w.Code)
		s.NoError(err)
		expected := "{\"message\":\"bad request: unexpected EOF\"}"
		s.Equal(expected, w.Body.String())
	})

	s.Run("When the payload has invalid type of value", func() {
		s.cu.On("Insert", mock.Anything, mock.Anything).Return(nil)
		req, err := http.NewRequest("POST", "/task/", strings.NewReader("{\"title\": 1,\"description\": \"td00002\"}"))
		s.NoError(err)
		w := httptest.NewRecorder()
		s.handler.InsertTask(w, req)
		s.Equal(http.StatusBadRequest, w.Code)
		s.NoError(err)
		expected := "{\"message\":\"bad request:. Wrong Type provided for field title\"}"
		s.Equal(expected, w.Body.String())
	})

	s.Run("When the payload with description empty value", func() {
		s.cu.On("Insert", mock.Anything, mock.Anything).Return(nil)
		req, err := http.NewRequest("POST", "/task/", strings.NewReader("{\"title\": \"title\",\"description\": \" \"}"))
		s.NoError(err)
		w := httptest.NewRecorder()
		s.handler.InsertTask(w, req)
		s.Equal(http.StatusBadRequest, w.Code)
		s.NoError(err)
		expected := "{\"message\":\"bad request: Requiered field description\"}"
		s.Equal(expected, w.Body.String())
	})

	s.Run("When the payload with title empty value", func() {
		s.cu.On("Insert", mock.Anything, mock.Anything).Return(nil)
		req, err := http.NewRequest("POST", "/task/", strings.NewReader("{\"title\": \"  \",\"description\": \"desc\"}"))
		s.NoError(err)
		w := httptest.NewRecorder()
		s.handler.InsertTask(w, req)
		s.Equal(http.StatusBadRequest, w.Code)
		s.NoError(err)
		expected := "{\"message\":\"bad request: Requiered field title\"}"
		s.Equal(expected, w.Body.String())
	})
}

func (s *SuiteTodo) TestUpdate() {
	s.Run("When the use case is succesful", func() {
		s.cu.On("Update", mock.Anything, "000000", mock.Anything).Return(nil)
		req, err := http.NewRequest("PUT", "/task/000000", strings.NewReader("{\"title\": \"t001\",\"description\": \"td00001\"}"))
		s.NoError(err)
		vars := map[string]string{"task_id": "000000"}
		req = mux.SetURLVars(req, vars)
		w := httptest.NewRecorder()
		s.handler.UpdateTask(w, req)
		s.Equal(http.StatusAccepted, w.Code)
		s.NoError(err)
		expected := "{\"data\":{\"id\":\"00000000-0000-0000-0000-000000000000\",\"title\":\"t001\",\"description\":\"td00001\"}}"
		s.Equal(expected, w.Body.String())
	})

	s.Run("When the use case return a generic error", func() {
		s.cu.On("Update", mock.Anything, "000002", mock.Anything).
			Return(errors.New("G error"))
		req, err := http.NewRequest("PUT", "/task/000002", strings.NewReader("{\"title\": \"t001\",\"description\": \"td00001\"}"))
		s.NoError(err)
		vars := map[string]string{"task_id": "000002"}
		req = mux.SetURLVars(req, vars)
		w := httptest.NewRecorder()
		s.handler.UpdateTask(w, req)
		s.Equal(http.StatusBadRequest, w.Code)
		s.NoError(err)
		expected := "{\"message\":\"G error\"}"
		s.Equal(expected, w.Body.String())
	})

	s.Run("When the payload has a error", func() {
		s.cu.On("Update", mock.Anything, "000003", mock.Anything).Return(nil)
		req, err := http.NewRequest("PUT", "/task/000003", strings.NewReader("{\"title\": \"t002\",\"description\": \"td00002}"))
		s.NoError(err)
		w := httptest.NewRecorder()
		s.handler.UpdateTask(w, req)
		s.Equal(http.StatusBadRequest, w.Code)
		s.NoError(err)
		expected := "{\"message\":\"bad request: unexpected EOF\"}"
		s.Equal(expected, w.Body.String())
	})

	s.Run("When the payload with description empty value", func() {
		s.cu.On("Update", mock.Anything, "000003", mock.Anything).Return(nil)
		req, err := http.NewRequest("PUT", "/task/000003", strings.NewReader("{\"title\": \"title\",\"description\": \" \"}"))
		s.NoError(err)
		w := httptest.NewRecorder()
		s.handler.UpdateTask(w, req)
		s.Equal(http.StatusBadRequest, w.Code)
		s.NoError(err)
		expected := "{\"message\":\"bad request: Requiered field description\"}"
		s.Equal(expected, w.Body.String())
	})
}

func (s *SuiteTodo) TestDelete() {
	s.Run("When the use case is succesful", func() {
		s.cu.On("Delete", mock.Anything, "000").Return(nil)
		req, err := http.NewRequest("DELETE", "/task/000", strings.NewReader(""))
		s.NoError(err)
		vars := map[string]string{"task_id": "000"}
		req = mux.SetURLVars(req, vars)
		w := httptest.NewRecorder()
		s.handler.DeleteTask(w, req)
		s.Equal(http.StatusAccepted, w.Code)
		s.NoError(err)
		expected := "{}"
		s.Equal(expected, w.Body.String())
	})

	s.Run("When the use case is succesful", func() {
		s.cu.On("Delete", mock.Anything, "001").Return(errors.New("G error"))
		req, err := http.NewRequest("DELETE", "/task/001", strings.NewReader(""))
		s.NoError(err)
		vars := map[string]string{"task_id": "001"}
		req = mux.SetURLVars(req, vars)
		w := httptest.NewRecorder()
		s.handler.DeleteTask(w, req)
		s.Equal(http.StatusNotFound, w.Code)
		s.NoError(err)
		expected := "{\"message\":\"G error\"}"
		s.Equal(expected, w.Body.String())
	})
}

func (s *SuiteTodo) TestGetTask() {
	s.Run("When the use case is succesful", func() {
		task := domain.NewTask("title 01", "domain 01")
		s.cu.On("GetByID", mock.Anything, "01").Return(task, nil)
		req, err := http.NewRequest("GET", "/task/01", strings.NewReader(""))
		s.NoError(err)
		vars := map[string]string{"task_id": "01"}
		req = mux.SetURLVars(req, vars)
		w := httptest.NewRecorder()
		s.handler.GetTask(w, req)
		s.Equal(http.StatusOK, w.Code)
		s.NoError(err)
		expected := "{\"data\":{\"id\":\"00000000-0000-0000-0000-000000000000\",\"title\":\"title 01\",\"description\":\"domain 01\"}}"
		s.Equal(expected, w.Body.String())
	})

	s.Run("When the use case is succesful", func() {
		s.cu.On("GetByID", mock.Anything, "02").Return(nil, errors.New("G error"))
		req, err := http.NewRequest("GET", "/task/02", strings.NewReader(""))
		s.NoError(err)
		vars := map[string]string{"task_id": "02"}
		req = mux.SetURLVars(req, vars)
		w := httptest.NewRecorder()
		s.handler.GetTask(w, req)
		s.Equal(http.StatusNotFound, w.Code)
		s.NoError(err)
		expected := "{\"message\":\"G error\"}"
		s.Equal(expected, w.Body.String())
	})
}

func TestSuiteTodo(t *testing.T) {
	suite.Run(t, new(SuiteTodo))
}
