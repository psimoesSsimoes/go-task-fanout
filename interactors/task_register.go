package interactors

import (
	"context"
	"time"

	"github.com/psimoesSsimoes/go-task-fanout/models"
)

//TaskRegisterRepository required interface to get tasks from todo to doing and Process them
type TaskRegisterRepository interface {
	CreateSchema(ctx context.Context) error
	GetTask(ctx context.Context, action string, age time.Duration) (models.Task, error)
	GetSeveralTasks(ctx context.Context, action string, age time.Duration) ([]models.Task, error)
	MarkAsDone(ctx context.Context, task models.Task) error
	MarkSeveralAsDone(ctx context.Context, task []models.Task) error
}

//TaskRegisterUpdater to hold all external abstractions
type TaskRegisterUpdater struct {
	repository TaskRegisterRepository
}

//NewTaskRegisterUpdater factory method
func NewTaskRegisterUpdater(r TaskRegisterRepository) TaskRegisterUpdater {
	return TaskRegisterUpdater{r}
}

//StartTask retrieves a task from todo position, changing it to doing position
func (i *TaskRegisterUpdater) StartTask(ctx context.Context, action string, age time.Duration) (models.Task, error) {

	return i.repository.GetTask(ctx, action, age)
}

//CompleteTask task marks a task as completed, removing it from doing position
func (i *TaskRegisterUpdater) CompleteTask(ctx context.Context, task models.Task) error {

	return i.repository.MarkAsDone(ctx, task)
}

//StartSeveralTask retrieves a task from todo position, changing it to doing position
func (i *TaskRegisterUpdater) StartSeveralTask(ctx context.Context, action string, age time.Duration) ([]models.Task, error) {

	return i.repository.GetSeveralTasks(ctx, action, age)
}

//CompleteSeveralTasks task marks a task as completed, removing it from doing position
func (i *TaskRegisterUpdater) CompleteSeveralTasks(ctx context.Context, tasks []models.Task) error {

	return i.repository.MarkSeveralAsDone(ctx, tasks)
}

//Init task Creates database schema
func (i *TaskRegisterUpdater) Init(ctx context.Context) error {

	return i.repository.CreateSchema(ctx)
}
