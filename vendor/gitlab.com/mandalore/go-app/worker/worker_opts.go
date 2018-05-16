package worker

import log "gitlab.com/vredens/go-logger"
import "time"

// WithMaxPerSecond allows you to configure how much the bucket leaks per second (see WithTick for configuring how often the bucket leaks).
func WithMaxPerSecond(max int) Option {
	return func(w *Worker) {
		if max > 0 {
			w.maxPerSecond = max
		}
	}
}

// SetMaxPerSecond updates the worker's maximum tasks processed per second.
func (w *Worker) SetMaxPerSecond(max int) {
	if max > 0 {
		w.maxPerSecond = max
	}
}

// WithTick allows you to configure how often the bucket leaks in Miliseconds. A value of 100 means the bucket leaks every 100 ms.
func WithTick(tick int) Option {
	return func(w *Worker) {
		if tick > 0 {
			w.tick = time.Millisecond * time.Duration(tick)
		}
	}
}

// SetTick updates the worker's tick interval. If the tick interval is 0 or less this method silently ignores the request.
func (w *Worker) SetTick(delay int) {
	if delay > 0 {
		w.tick = time.Millisecond * time.Duration(delay)
	}
}

// WithProcessDelay allows configuring how old a task must be in order to be processed, this allows backing off tasks until a certain
// time has passed in order to de-duplicate tasks in cases of bursts. Process delay is in miliseconds.
func WithProcessDelay(delay int) Option {
	return func(w *Worker) {
		w.processDelay = time.Duration(delay) * time.Millisecond
	}
}

// SetProcessDelay updated the process delay for the worker to the provided delay value. The delay must be higher than 0, else this method silently ignores the request.
func (w *Worker) SetProcessDelay(delay int) {
	if delay > 0 {
		w.processDelay = time.Duration(delay) * time.Millisecond
	}
}

// WithWorkHandler this will configure the handler for each work job. This is a required option (insanity!).
func WithWorkHandler(f WorkHandler) Option {
	return func(w *Worker) {
		w.handler = f
	}
}

// SetWorkHandler updates the worker's task handler.
func (w *Worker) SetWorkHandler(f WorkHandler) {
	w.handler = f
}

// WithQueueSize allows you to create a configuration for the max size of the queue. Once the queue is full all job requests will block.
func WithQueueSize(size int) Option {
	return func(w *Worker) {
		if size > 0 {
			w.queueSize = size
		}
	}
}

// WithLogger ...
func WithLogger(l log.Logger) Option {
	return func(w *Worker) {
		w.log = l
	}
}
