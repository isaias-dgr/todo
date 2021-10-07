package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/isaias-dgr/todo/src/domain"
	"go.uber.org/zap"
)

type TaskHandler struct {
	TuseCase domain.TaskUseCase
	L        *zap.SugaredLogger
}

func NewTaskHandler(taskUseCase domain.TaskUseCase, logger *zap.SugaredLogger) {
	handler := &TaskHandler{
		TuseCase: taskUseCase,
		L:        logger,
	}

	r := mux.NewRouter()
	r.HandleFunc("/task/", handler.FetchTasks).Methods("GET")
	r.HandleFunc("/task/", handler.InsertTask).Methods("POST")
	r.HandleFunc("/task/{task_id}/", handler.GetTask).Methods("GET")
	r.HandleFunc("/task/{task_id}/", handler.UpdateTask).Methods("PUT")
	r.HandleFunc("/task/{task_id}/", handler.DeleteTask).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func (t *TaskHandler) FetchTasks(w http.ResponseWriter, r *http.Request) {
	t.L.Infow("Fetch", "url", r.URL, "method", r.Method)
	filter := domain.NewFilter(r.URL.Query())
	tasks, err := t.TuseCase.Fetch(r.Context(), filter)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	makeResponse(w, http.StatusOK, tasks.Data, filter, tasks.Total)
}

func (t *TaskHandler) InsertTask(w http.ResponseWriter, r *http.Request) {
	t.L.Infow("Insert", "url", r.URL, "method", r.Method)
	var task domain.Task
	if err := t.DecoderBody(r.Body, &task); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validate(&task); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := t.TuseCase.Insert(r.Context(), &task); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	makeResponse(w, http.StatusAccepted, task, nil, 0)
}

func (t *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	t.L.Infow("Update", "url", r.URL, "method", r.Method)
	var task domain.Task

	if err := t.DecoderBody(r.Body, &task); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validate(&task); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	vars := mux.Vars(r)
	err := t.TuseCase.Update(r.Context(), vars["task_id"], &task)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	makeResponse(w, http.StatusAccepted, task, nil, 0)
}

func (t *TaskHandler) DecoderBody(b io.ReadCloser, ta *domain.Task) error {
	var unmarshalErr *json.UnmarshalTypeError
	decoder := json.NewDecoder(b)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(ta)
	if err != nil {
		t.L.Error("Bad Request. %s", err.Error())
		if errors.As(err, &unmarshalErr) {
			return fmt.Errorf("bad request:. Wrong Type provided for field %s", unmarshalErr.Field)
		} else {
			return fmt.Errorf("bad request: %s", err.Error())
		}
	}
	return nil
}

func validate(t *domain.Task) error {
	if strings.TrimSpace(t.Title) == "" {
		return errors.New("bad request: Requiered field title")
	}

	if strings.TrimSpace(t.Description) == "" {
		return errors.New("bad request: Requiered field description")
	}
	return nil
}

func (t *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	t.L.Infow("Get by uuid", "url", r.URL, "method", r.Method)
	vars := mux.Vars(r)
	task, err := t.TuseCase.GetByID(r.Context(), vars["task_id"])
	if err != nil {
		errorResponse(w, http.StatusNotFound, err.Error())
		return
	}
	makeResponse(w, http.StatusOK, task, nil, 0)
}

func (t *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	t.L.Infow("Delete", "url", r.URL, "method", r.Method)
	vars := mux.Vars(r)
	err := t.TuseCase.Delete(r.Context(), vars["task_id"])
	if err != nil {
		errorResponse(w, http.StatusNotFound, err.Error())
		return
	}
	makeResponse(w, http.StatusAccepted, nil, nil, 0)
}

func makeResponse(w http.ResponseWriter, code int, body interface{}, filter *domain.Filter, total int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	resp := domain.NewResponse(body, total, filter, "")
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)
}

func errorResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	resp := make(map[string]string)
	resp["message"] = message
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)
}
