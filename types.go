package taskworker

import (
	"time"
)

// Task is a work task
type Task struct {
	ID        int
	TaskID    string
	Data      interface{}
	Action    string
	CreatedAt time.Time
	StartedAt time.Time
}

// TaskStorage manages tasks.
type TaskStorage interface {
	// Create stores a task for processing.
	Create(task *Task) error
	// Get returns the next Task for command which is in the 'todo' state.
	Get(command string, age time.Duration) (*Task, error)
	// GetBatch returns the next N Tasks for command which is in the 'todo' state.
	GetBatch(command string, age time.Duration, n int) ([]*Task, error)
	// Retry marks all tasks for command in the state 'done' and older than 'age' back to the 'todo' state.
	Retry(command string, age time.Duration) error
	// Cleanup removes all tasks for command in the 'done' older than 'age'
	Cleanup(command string, age time.Duration) error
	// Complete marks a task as complete.
	Complete(task *Task) error
	// Fail fails the task and increase retries count.
	Fail(task *Task, reason string) error
}

// Logger as the name says, it do logging
type Logger interface {
}
