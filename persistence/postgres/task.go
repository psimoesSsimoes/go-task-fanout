package postgres

import (
	"database/sql"
	"time"

	"fmt"

	"encoding/json"

	"github.com/pkg/errors"
)

// TaskStorage manages tasks.
type TaskStorage struct {
	conn       *sql.DB
	subject    string
	todoTable  string
	doingTable string
}

// NewTaskStorage creates a task storage
func NewTaskStorage(conn *sql.DB, subject string) *TaskStorage {
	return &TaskStorage{
		conn:       conn,
		subject:    subject,
		todoTable:  fmt.Sprintf("%s_todo", subject),
		doingTable: fmt.Sprintf("%s_doing", subject),
	}
}

// Init prepares the storage, if needed, to manage task to a specific subject.
func (s *TaskStorage) Init() error {
	_, err := s.conn.Exec(`
		BEGIN;

		CREATE schema IF NOT EXISTS workqueue;

		CREATE TABLE IF NOT EXISTS workqueue.` + s.todoTable + `
		(
			id SERIAL PRIMARY KEY,
			task_id TEXT NOT NULL,
			action TEXT NOT NULL,
			data JSONB DEFAULT '{}',
			created_at TIMESTAMP DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS workqueue.` + s.doingTable + `
		(
			id SERIAL PRIMARY KEY,
			task_id TEXT NOT NULL,
			action TEXT NOT NULL,
			data JSONB DEFAULT '{}',
			created_at TIMESTAMP DEFAULT NOW(),
			started_at TIMESTAMP DEFAULT NOW()
		);

		COMMIT;
	`)

	if err != nil {
		return errors.Wrap(err, "error occurred creating the task")
	}

	return nil
}

// Create stores a task for processing.
func (s *TaskStorage) Create(task *taskworker.Task) error {
	data, err := json.Marshal(task.Data)
	if err != nil {
		return errors.Wrap(err, "error occurred creating the task")

	}

	_, err = s.conn.Exec(`
		INSERT INTO workqueue.`+s.todoTable+`(task_id, action, data)
		SELECT $1, $2, $3
		FROM   workqueue.`+s.todoTable+`
		WHERE  task_id = $1
			AND action = $2
		HAVING count(1) = 0;
	`, task.TaskID, task.Action, data)

	if err != nil {
		return errors.Wrap(err, "error occurred creating the task")
	}

	return nil
}

// Get returns the next Task for command which is in the 'todo' state.
func (s *TaskStorage) Get(action string, age time.Duration) (*taskworker.Task, error) {
	task := &taskworker.Task{}

	row := s.conn.QueryRow(`
		WITH moved_rows AS (
			DELETE FROM workqueue.`+s.todoTable+`
			WHERE id IN (
				SELECT id
				FROM workqueue.`+s.todoTable+`
				WHERE action = $1
					AND age(current_timestamp, created_at) > $2
				ORDER BY created_at ASC
				LIMIT 1
			)
			RETURNING *
		)
		INSERT INTO workqueue.`+s.doingTable+`(id, task_id, action, data)
		SELECT id, task_id, action, data
		FROM moved_rows
		RETURNING id, task_id, action, data, created_at, started_at;
	`, action, age)

	if err := row.Scan(
		&task.ID,
		&task.TaskID,
		&task.Action,
		&task.Data,
		&task.CreatedAt,
		&task.StartedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, errors.Wrap(err, "error occurred getting tasks")
	}

	return task, nil
}

// GetBatch returns the next N Tasks for command which is in the 'todo' state.
func (s *TaskStorage) GetBatch(command string, age time.Duration, n int) ([]*taskworker.Task, error) {
	rows, err := s.conn.Query(`
		WITH moved_rows AS (
			DELETE FROM workqueue.`+s.todoTable+`
			WHERE id IN (
				SELECT id
				FROM workqueue.`+s.todoTable+`
				WHERE action = $1
					AND age(current_timestamp, created_at) > $2
				ORDER BY created_at ASC
				LIMIT $3
			)
			RETURNING *
		)
		INSERT INTO workqueue.`+s.doingTable+`(id, task_id, action, data)
		SELECT id, task_id, action, data
		FROM moved_rows
		RETURNING id, task_id, action, data, created_at, started_at;
	`, command, age, n)

	if err != nil {
		return nil, errors.Wrap(err, "error occurred getting tasks")
	}

	defer rows.Close()

	tasks := make([]*taskworker.Task, 0)

	for rows.Next() {
		task := &taskworker.Task{}

		if err := rows.Scan(
			&task.ID,
			&task.TaskID,
			&task.Action,
			&task.Data,
			&task.CreatedAt,
			&task.StartedAt,
		); err != nil {
			return nil, errors.Wrap(err, "error occurred getting tasks")
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// Retry marks all tasks for command in the state 'done' and older than 'age' back to the 'todo' state.
func (s *TaskStorage) Retry(command string, age time.Duration) error {
	return nil
}

// Cleanup removes all tasks for command in the 'done' older than 'age'
func (s *TaskStorage) Cleanup(command string, age time.Duration) error {
	return nil
}

// Complete marks a task as complete.
func (s *TaskStorage) Complete(task *taskworker.Task) error {
	_, err := s.conn.Exec(`
		DELETE FROM workqueue.`+s.doingTable+`
		WHERE id = $1
	`, task.ID)

	if err != nil {
		return errors.Wrap(err, "error occurred completing the task")
	}

	return nil
}

// Fail fails the task and increase retries count.
func (s *TaskStorage) Fail(task *taskworker.Task, reason string) error {
	/*_, err := s.conn.Exec(`
		UPDATE taskworker.tasks
		SET state = CASE WHEN retries + 1 < 5 THEN $2 ELSE $3 END,
			retries = CASE WHEN retries < 5 THEN retries + 1 ELSE retries END,
			updated_at = current_timestamp,
			fail_reason = $4
		WHERE id = $1
	`, task.ID, TaskStateToDo, TaskStateFailed, reason)

	if err != nil {
		return errors.Wrap(err, "error occurred failing the task")
	}*/

	return nil
}
