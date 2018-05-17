package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/psimoesSsimoes/go-task-fanout/models"
)

// RegisterRepository type holds the postgres connection pool
type RegisterRepository struct {
	pool *sql.DB
}

// NewRegisterRepository factory method to create a new repository instance
func NewRegisterRepository(pool *sql.DB) RegisterRepository {
	return RegisterRepository{pool}
}

func (r *RegisterRepository) GetTask(ctx context.Context, action string, age time.Duration) (models.Task, error) {
	return models.Task{}, nil
}

func (r *RegisterRepository) GetSeveralTask(ctx context.Context, action string, age time.Duration) ([]models.Task, error) {
	return []models.Task{{}}, nil
}
func (r *RegisterRepository) MarkAsDone(ctx context.Context, action string, age time.Duration) error {
	return models.Task{}, nil
}

func (r *RegisterRepository) MarkSeveralAsDone(ctx context.Context, action string, age time.Duration) error {
	return models.Task{}, nil
}
