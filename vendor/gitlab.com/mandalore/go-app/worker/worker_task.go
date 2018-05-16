package worker

import "time"

type Task struct {
	id   string
	data interface{}
	ts   time.Time
}

func NewTask(id string, data interface{}) *Task {
	return &Task{
		id:   id,
		ts:   time.Now(),
		data: data,
	}
}

func (t *Task) Age() time.Duration {
	return time.Since(t.ts)
}
