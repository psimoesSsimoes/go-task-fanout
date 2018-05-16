package interactors

import (
	"context"

	"github.com/psimoesSsimoes/go-task-fanout/models"
)

//TaskDispatcherRepository required interface to store tasks on todo table
type TaskDispatcherRepository interface {
	CreateTask(ctx context.Context, task models.Task) error
}

//TaskDispatcherUpdater to holds all external abstractions
type TaskDispatcherUpdater struct {
	repository TaskDispatcherRepository
}

//NewTaskDispatcher factory method
func NewTaskDispatcher(r TaskDispatcherRepository) TaskDispatcherUpdater {
	return TaskDispatcherUpdater{r}
}

//Process stores a new Task in todo using repo
func (i *TaskDispatcherUpdater) Process(ctx context.Context, task models.Task) error {

	return i.repository.CreateTask(ctx, task)
}
