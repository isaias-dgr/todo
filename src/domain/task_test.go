package domain_test

import (
	"testing"

	"github.com/isaias-dgr/todo/src/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewTask(t *testing.T) {
	assert := assert.New(t)
	title := "title"
	description := "description"
	task := domain.NewTask(title, description)
	assert.Equal(task.Title, title, "The title is incorrect")
	assert.Equal(task.Description, description, "The title is incorrect")
}

func TestNewTasks(t *testing.T) {
	assert := assert.New(t)
	total_task := 3
	tasks := []*domain.Task{
		domain.NewTask("title 2", "description 2"),
		domain.NewTask("title 3", "description 3"),
		domain.NewTask("title 4", "description 4"),
	}
	ts := domain.NewTasks(tasks, total_task)
	assert.Equal(ts.Total, total_task)
}
