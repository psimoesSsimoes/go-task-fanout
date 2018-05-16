package taskworker

import (
	"github.com/pkg/errors"
)

// Dispatcher is a task dispatcher to a specific command
type Dispatcher struct {
	storage TaskStorage
}

// NewDispatcher creates a new dispatcher
func NewDispatcher(storage TaskStorage) *Dispatcher {
	return &Dispatcher{
		storage: storage,
	}
}

// Process adds a task
func (d *Dispatcher) Process(command string, id string, data interface{}) error {
	task := &Task{
		TaskID: id,
		Data:   data,
		Action: command,
	}

	if err := d.storage.Create(task); err != nil {
		return errors.Wrap(err, "failed to create task")
	}
	return nil
}
