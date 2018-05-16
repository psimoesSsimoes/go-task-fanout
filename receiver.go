package taskworker

import (
	"sync"
	"time"

	"fmt"

	"github.com/pkg/errors"
	"gitlab.com/mandalore/go-app/app"
	"gitlab.com/vredens/go-logger"
)

const (
	// StateReady is the state a worker is in after being initialized.
	StateReady = iota
	// StateRunning is when the worker has started successfully.
	StateRunning
	// StateProcessing is when the worker has processing tasks.
	StateProcessing
	// StateStopping is when a Stop command has been requested but the worker hasn't stopped yet. A worker in this state refuses to add more tasks to the queue.
	StateStopping
	// StateStopped is when a worker has been successfully stopped and no tasks remain on the queue.
	StateStopped
	// StateDying is set when a Kill command is issued. This prevents further tasks from being added to the queue while the fatal process is under way.
	StateDying
	// StateDead happens when a worker has been killed.
	StateDead
	// StateError happens when a worker in an error state. A worker in this state can only be destroyed.
	StateError
)

var workerStateString = map[int]string{
	StateReady:      "READY",
	StateRunning:    "RUNNING",
	StateProcessing: "PROCESSING",
	StateStopping:   "STOPPING",
	StateStopped:    "STOPPED",
	StateDying:      "DYING",
	StateDead:       "DEAD",
	StateError:      "ERROR",
}

// TaskHandler is the function signature for any function capable of handling a task.
// Both the task TaskID and task data are provided to the handler. The task TaskID can and
// should be used to identify task types by the way of a prefix or bit mask.
type TaskHandler func(*Task) error

// ReceiverOption is the abstract functional-parameter type used for worker configuration.
type ReceiverOption func(*Receiver)

// Receiver is the task e handler
type Receiver struct {
	mux       *sync.Mutex
	state     int
	batchSize int
	age       time.Duration
	tick      time.Duration
	control   chan bool
	command   string
	storage   TaskStorage
	handler   TaskHandler
	logger    logger.Logger
}

// WithLogger allows you to configure the logger.
func WithLogger(logger logger.Logger) ReceiverOption {
	return func(r *Receiver) {
		if logger != nil {
			r.logger = logger
		}
	}
}

// WithTick allows you to configure how often the bucket leaks in Miliseconds. A value of 100 means the bucket leaks every 100 ms.
func WithTick(tick int) ReceiverOption {
	return func(r *Receiver) {
		if tick > 0 {
			r.tick = time.Millisecond * time.Duration(tick)
		}
	}
}

// WithWorkHandler this will configure the handler for each work job. This is a required option (insanity!).
func WithWorkHandler(f TaskHandler) ReceiverOption {
	return func(r *Receiver) {
		r.handler = f
	}
}

// WithBatchSize allows you to create a configuration for the max size of the tasks. Once the is has no tasks, receiver will ask for more.
func WithBatchSize(size int) ReceiverOption {
	return func(r *Receiver) {
		if size > 0 {
			r.batchSize = size
		}
	}
}

// WithTaskAge allows you to set the task age allowed to be processed
func WithTaskAge(age time.Duration) ReceiverOption {
	return func(r *Receiver) {
		r.age = age
	}
}

// NewReceiver creates a new receiver
func NewReceiver(storage TaskStorage, command string, opts ...ReceiverOption) *Receiver {
	r := &Receiver{
		storage:   storage,
		command:   command,
		state:     StateReady,
		tick:      time.Second,
		batchSize: 1000,
		age:       time.Duration(24 * time.Hour),
		control:   make(chan bool),
		mux:       &sync.Mutex{},
		logger:    logger.SpawnMute(),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (r *Receiver) setState(state int) error {
	r.mux.Lock()

	if state == StateError {
		r.state = StateError
		r.mux.Unlock()

		return nil
	}

	var validTransition bool

	switch r.state {
	case StateReady:
		switch state {
		case StateRunning, StateStopping:
			validTransition = true
		}
	case StateRunning:
		switch state {
		case StateProcessing, StateStopping, StateDying:
			validTransition = true
		}
	case StateProcessing:
		switch state {
		case StateRunning, StateStopping, StateDying:
			validTransition = true
		}
	case StateStopping:
		switch state {
		case StateStopped:
			validTransition = true
		}
	case StateStopped:
		switch state {
		case StateRunning:
			validTransition = true
		}
	case StateDying:
		switch state {
		case StateDead:
			validTransition = true
		}
	}

	if !validTransition {
		r.mux.Unlock()

		return fmt.Errorf("invalid state transition [current:%s][next:%s]", workerStateString[r.state], workerStateString[state])
	}

	r.state = state
	r.mux.Unlock()

	return nil
}

// Start starts the process.
func (r *Receiver) Start() error {
	if r.handler == nil {
		return errors.New("no task handler was set")
	}

	if err := r.setState(StateRunning); err != nil {
		return err
	}

	for {
		select {
		case <-time.After(r.tick):
			if r.state == StateStopping {
				return nil
			}

			if err := r.setState(StateProcessing); err != nil {
				r.logger.WithData(app.KV{"cause": app.StringifyError(err)}).Info("failed to set state")

				continue
			}

			tasks, err := r.storage.GetBatch(r.command, r.age, r.batchSize)
			if err != nil {
				r.logger.WithData(app.KV{"cause": app.StringifyError(err)}).Warn("failed to get task batch")
				if err := r.setState(StateRunning); err != nil {
					r.logger.WithData(app.KV{"cause": app.StringifyError(err)}).Info("failed to set state")

					continue
				}

				continue
			}

			r.processBatch(tasks)

			if err := r.setState(StateRunning); err != nil {
				r.logger.WithData(app.KV{"cause": app.StringifyError(err)}).Info("failed to set state")

				continue
			}
		case <-r.control:
			if r.state == StateDying {
				r.setState(StateDead)

				return nil
			}
		}
	}
}

func (r *Receiver) processBatch(tasks []*Task) {
	for _, task := range tasks {
		if err := r.processTask(task); err != nil {
			r.logger.WithData(app.KV{"task_id": task.TaskID}).Info("failed to process task")

		}
	}
}

func (r *Receiver) processTask(task *Task) error {
	if err := r.handler(task); err != nil {
		if err := r.storage.Fail(task, err.Error()); err != nil {
			return errors.Wrap(err, "failed to mark task as failed")
		}
	}

	if err := r.storage.Complete(task); err != nil {
		return errors.Wrap(err, "failed to mark task as completed")
	}

	return nil
}

// Stop stops the process
func (r *Receiver) Stop() error {
	if err := r.setState(StateStopping); err != nil {
		return err
	}
	if r.state == StateProcessing {
		r.logger.Warn("stopping worker with tasks in queue")
	}

	return nil
}

// Kill kills the process
func (r *Receiver) Kill() error {
	close(r.control)

	return nil
}
