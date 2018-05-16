package models

import (
	"time"

	"github.com/oklog/ulid"
)

//Task represents a work task
type Task struct {
	ID        ulid.ULID
	TaskID    string
	Data      interface{}
	Action    string
	CreatedAt time.Time
	StartedAt time.Time
}

//TaskOption helper type to define firt order optional params
type TaskOption func(t *Task)

//NewTask creates a new task
func NewTask(id ulid.ULID, taskid, action string, data interface{}, optionFuncs ...TaskOption) Task {

	task := Task{
		ID:     id,
		TaskID: taskid,
		Data:   data,
		Action: action,
	}

	for _, optionFunc := range optionFuncs {
		optionFunc(&task)
	}

	return task
}

// WithCreatedAt option first order function to use as composable on factory methods
func WithCreatedAt(t time.Time) TaskOption {
	return func(task *Task) {
		task.CreatedAt = t
	}
}

// WithStartedAt option first order function to use as composable on factory methods
func WithStartedAt(t time.Time) TaskOption {
	return func(task *Task) {
		task.StartedAt = t
	}
}
