package infrastructure

import (
	"github.com/psimoesSsimoes/go-task-fanout/interactors"
)

// Infrastructure interface to define infrastruture factories for later injection
type Infrastructure interface {
	TaskDispatcherRepository() interactors.TaskDispatcherRepository
	TaskRegisterRepository() interactors.TaskRegisterRepository
}
