package worker

import (
	"sync"
	"time"

	"gitlab.com/mandalore/go-app/app"
	log "gitlab.com/vredens/go-logger"
	"gitlab.com/vredens/go-structs"
)

// WorkHandler is the function signature for any function capable of handling a task.
// Both the task ID and task data are provided to the handler. The task ID can and
// should be used to identify task types by the way of a prefix or bit mask.
type WorkHandler func(id string, data interface{}) error

// Option is the abstract functional-parameter type used for worker configuration.
type Option func(*Worker)

const (
	// StateNone represents a worker with no state which means it hasn't been initialized yet.
	StateNone int = iota
	// StateReady is the state a worker is in after being initialized.
	StateReady
	// StateRunning is when the worker has started successfully.
	StateRunning
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
	StateNone:     "EMBRIONIC",
	StateReady:    "READY",
	StateRunning:  "RUNNING",
	StateStopping: "STOPPING",
	StateStopped:  "STOPPED",
	StateDying:    "DYING",
	StateDead:     "DEAD",
	StateError:    "ERROR",
}

// Worker is a generic implementation of a task worker providing processing rate limits and task process delay. Tasks are indexed by an ID so the same task is not performed twice.
type Worker struct {
	queue        *structs.Queue
	mux          *sync.Mutex
	maxPerSecond int
	queueSize    int
	tick         time.Duration
	processDelay time.Duration
	control      chan bool
	state        int
	handler      WorkHandler
	err          app.Error
	log          log.Logger
}

// NewWorker creates a new Worker with the provided options.
func NewWorker(opts ...Option) *Worker {
	worker := &Worker{
		maxPerSecond: 0,
		tick:         time.Second,
		queueSize:    10000,
		processDelay: time.Second,
		control:      make(chan bool),
		state:        StateNone,
		mux:          &sync.Mutex{},
		handler:      nil,
		log:          log.Spawn(log.WithTags("worker")),
	}

	for _, opt := range opts {
		opt(worker)
	}

	worker.queue = structs.NewQueue(worker.queueSize)

	if err := worker.setState(StateReady); err != nil {
		panic(err)
	}

	return worker
}

// Error returns the worker error which is only different from nil if the worker is in the StateError.
// This only happens when an error occurs during worker execution such as failure to
func (w *Worker) Error() error {
	return w.err
}

func (w *Worker) setState(state int) error {
	if state == StateError {
		w.state = StateError
		return nil
	}

	var validTransition bool

	switch w.state {
	case StateNone:
		switch state {
		case StateReady:
			validTransition = true
		}
	case StateReady:
		switch state {
		case StateRunning, StateStopping:
			validTransition = true
		}
	case StateRunning:
		switch state {
		case StateStopping, StateDying:
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
		return app.NewErrorf(app.ErrorDevPoo, nil, "invalid state transition [current:%s][next:%s]", workerStateString[w.state], workerStateString[state])
	}

	w.log.Infof("changing state [from:%s][to:%s]", workerStateString[w.state], workerStateString[state])
	w.state = state
	return nil
}

// SetHandler allows setting a work handler even after starting the worker.
// DEPRECATED method. Please use the WithWorkHandler functional parameter.
func (w *Worker) SetHandler(f WorkHandler) {
	w.mux.Lock()
	w.handler = f
	w.mux.Unlock()
}

// Start the worker.
func (w *Worker) Start() error {
	w.mux.Lock()
	if w.handler == nil {
		w.mux.Unlock()
		return app.NewError(app.ErrorDevPoo, "no task handler was set", nil)
	}
	if err := w.setState(StateRunning); err != nil {
		w.mux.Unlock()
		return err
	}
	w.mux.Unlock()

	// TODO: use golang.org/x/time/rate Limiter.WaitN instead, combines max per second with tick

	for {
		select {
		case <-time.After(w.tick):
			if w.queue.Size() > 0 {
				size := w.queue.Size()
				w.log.Debugf("process batch starting [size:%d]", w.queue.Size())
				w.processBatch()
				w.log.Debugf("process batch done [size:%d]", w.queue.Size())
				app.StatsAverageAdd("worker-process-batch", float64(size-w.queue.Size()))
			} else {
				w.mux.Lock()
				if w.state == StateStopping {
					w.setState(StateStopped)
					w.mux.Unlock()
					return nil
				}
				w.mux.Unlock()
			}
		case <-w.control:
			w.mux.Lock()
			if w.state == StateDying {
				w.setState(StateDead)
				w.mux.Unlock()
				return nil
			}
			w.mux.Unlock()
		}
	}
}

// Stop blocks any more tasks from entering the queue but waits for the queue to be empty.
func (w *Worker) Stop() error {
	w.mux.Lock()
	if err := w.setState(StateStopping); err != nil {
		w.mux.Unlock()
		return err
	}
	if w.queue.Size() > 0 {
		w.log.Warnf("stopping worker with tasks in queue [size:%d]", w.queue.Size())
	}
	w.mux.Unlock()

	return nil
}

// Kill works similar to stop but terminates with a non-empty queue resulting in lost tasks.
func (w *Worker) Kill() error {
	w.mux.Lock()
	if err := w.setState(StateDying); err != nil {
		w.mux.Unlock()
		return err
	}
	if w.queue.Size() > 0 {
		w.log.Errorf("terminating worker with unhandled tasks [size:%d]", w.queue.Size())
	}
	w.mux.Unlock()
	w.control <- true

	return nil
}

// Process ...
func (w *Worker) Process(taskID string, data interface{}) error {
	w.mux.Lock()
	if w.state != StateRunning && w.state != StateReady {
		err := app.NewErrorf(app.ErrorConflict, nil, "can not add tasks to a [%s] worker", workerStateString[w.state])
		w.mux.Unlock()
		return err
	}
	w.mux.Unlock()

	task := NewTask(taskID, data)
	return w.queue.Add(task.id, task)
}

func (w *Worker) processBatch() {
	cont := w.processNextTask()
	cnt := 1
	for cont && (w.maxPerSecond == 0 || float64(cnt) < float64(w.maxPerSecond)*w.tick.Seconds()) {
		cont = w.processNextTask()
		cnt++
	}
}

func (w *Worker) processNextTask() bool {
	var task *Task
	if tmp := w.queue.Pop(); tmp != nil {
		task = tmp.(*Task)
	} else {
		return false
	}

	if w.processDelay > 0 && task.Age() < w.processDelay {
		if err := w.queue.Add(task.id, task); err != nil {
			w.log.Errorf("unexpected error re-adding a task to the queue; %s", app.StringifyError(err))
		}
		return false
	}
	if err := w.handler(task.id, task.data); err != nil {
		w.log.Errorf("task handler failed; %s", app.StringifyError(err))
		return false
	}
	return true
}
