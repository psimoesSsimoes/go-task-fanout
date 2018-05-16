package interactors

import (
	"context"

	"github.com/psimoesSsimoes/go-task-fanout/models"
)

//TaskDispatcherRepository required interface to store tasks
type TaskDispatcherRepository interface {
	Create(ctx context.Context, task models.Task) error
}

//TaskDispatcherUpdater to holds all external abstractions
type TaskDispatcherUpdater struct {
	repository TaskDispatcherRepository
}

//NewTaskDispatcher factory method
func NewTaskDispatcher(r TaskDispatcherRepository) TaskDispatcherUpdater {
	return TaskDispatcherUpdater{r}
}

func (i *TaskDispatcherUpdater) Process(ctx context.Context, task models.Task) error {

	if err := i.repository.Create(ctx, task); err != nil {
		return err // err will come already wrapped
	}

	return nil
}
