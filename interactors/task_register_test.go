package interactors

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/psimoesSsimoes/go-task-fanout/models"
	mocks "github.com/psimoesSsimoes/go-task-fanout/tests/mocks/interactors"
	"github.com/stretchr/testify/mock"
)

func TestRegisterGetOnSuccess(t *testing.T) {
	RegisterTestingT(t)

	repository := &mocks.TaskRegisterRepository{}
	interactor := NewTaskRegisterUpdater(repository)

	// Mock behaviour must be defined before the actual call
	repository.On("GetTask", mock.Anything, "action", time.Duration(10)).Return(_generateSameTask(), nil)

	task, err := interactor.StartTask(context.TODO(), "action", time.Duration(10))

	Expect(err).ToNot(HaveOccurred(), "should not return an error")
	Expect(repository.AssertExpectations(t)).To(BeTrue(), "all methods where called")
	Expect(task).To(Equal(_generateSameTask()), "task are the same")
}

func TestRegisterGetOnFailure(t *testing.T) {
	RegisterTestingT(t)

	repository := &mocks.TaskRegisterRepository{}
	interactor := NewTaskRegisterUpdater(repository)

	rError := errors.New("booom")
	// Mock behaviour must be defined before the actual call
	repository.On("GetTask", mock.Anything, "action", time.Duration(10)).Return(models.Task{}, rError)

	_, err := interactor.StartTask(context.TODO(), "action", time.Duration(10))

	Expect(err).To(HaveOccurred(), "should return an error")
	Expect(repository.AssertExpectations(t)).To(BeTrue(), "all methods where called")
	Expect(err).To(Equal(rError), "error is the same")
}

func TestRegisterMarkAsDoneSuccess(t *testing.T) {
	RegisterTestingT(t)

	repository := &mocks.TaskRegisterRepository{}
	interactor := NewTaskRegisterUpdater(repository)

	// Mock behaviour must be defined before the actual call
	repository.On("MarkAsDone", mock.Anything, _generateSameTask()).Return(nil)

	err := interactor.CompleteTask(context.TODO(), _generateSameTask())

	Expect(err).ToNot(HaveOccurred(), "should not return an error")
	Expect(repository.AssertExpectations(t)).To(BeTrue(), "all methods where called")
}

func TestRegisterMarkAsDoneOnFailure(t *testing.T) {
	RegisterTestingT(t)

	repository := &mocks.TaskRegisterRepository{}
	interactor := NewTaskRegisterUpdater(repository)

	rError := errors.New("booom")
	// Mock behaviour must be defined before the actual call
	repository.On("MarkAsDone", mock.Anything, _generateSameTask()).Return(rError)

	err := interactor.CompleteTask(context.TODO(), _generateSameTask())

	Expect(err).To(HaveOccurred(), "should return an error")
	Expect(repository.AssertExpectations(t)).To(BeTrue(), "all methods where called")
	Expect(err).To(Equal(rError), "error is the same")
}

// func _generateSameTask() models.Task {
// 	var i interface{}
// 	return models.NewTask(
// 		generateSameUlid(),
// 		"atask",
// 		"action",
// 		i,
// 	)
// }
