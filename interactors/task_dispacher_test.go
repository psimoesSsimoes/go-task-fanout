package interactors

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/oklog/ulid"
	. "github.com/onsi/gomega"
	"github.com/psimoesSsimoes/go-task-fanout/models"
	mocks "github.com/psimoesSsimoes/go-task-fanout/tests/mocks/interactors"
	"github.com/stretchr/testify/mock"
)

func TestDispatchInsertOnSuccess(t *testing.T) {
	RegisterTestingT(t)

	repository := &mocks.TaskDispatcherRepository{}
	interactor := NewTaskDispatcher(repository)

	// Mock behaviour must be defined before the actual call
	repository.On("CreateTask", mock.Anything, _generateSameTask()).Return(nil)

	err := interactor.Process(context.TODO(), _generateSameTask())

	Expect(err).ToNot(HaveOccurred(), "should not return an error")
	Expect(repository.AssertExpectations(t)).To(BeTrue(), "all methods where called")
}

func TestDispatchInsertOnFailure(t *testing.T) {
	RegisterTestingT(t)

	repository := &mocks.TaskDispatcherRepository{}
	interactor := NewTaskDispatcher(repository)

	rError := errors.New("boooooom")
	// Mock behaviour must be defined before the actual call
	repository.On("CreateTask", mock.Anything, _generateSameTask()).Return(rError)

	err := interactor.Process(context.TODO(), _generateSameTask())

	Expect(err).To(HaveOccurred(), "should return an error")
	Expect(err).To(Equal(rError), "should return same error")
	Expect(repository.AssertExpectations(t)).To(BeTrue(), "all methods where called")
}

func _generateSameTask() models.Task {
	var i interface{}
	return models.NewTask(
		generateSameUlid(),
		"atask",
		"action",
		i,
	)
}

func generateSameUlid() ulid.ULID {
	t := time.Unix(1000000, 0)
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	return ulid.MustNew(ulid.Timestamp(t), entropy)

}
