package interactors



import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	. "github.com/onsi/gomega"
	mocks "github.com/psimoesSsimoes/go-task-fanout/tests/mocks/interactors"
)

func TestDispatchInsertOnSuccess(t *testing.T) {
	RegisterTestingT(t)

	repository := &mocks.TaskDispatcherRepository{}
	interactor := NewTaskDispatcher(repository)

	// Mock behaviour must be defined before the actual call
	repository.On("Create", mock.Anything,_generateSameTask()).Return(nil)

	err := interactor.Process(context.TODO(), _generateSameTask())

	Expect(err).ToNot(HaveOccurred(), "should not return an error")
	Expect(repository.AssertExpectations(t)).To(BeTrue(), "all methods where called")
}

func TestDispatchInsertOnFailure(t *testing.T) {
	RegisterTestingT(t)

	repository := &mocks.TaskDispatcherRepository{}
	interactor := NewTaskDispatcher(repository)

	rError:= errors.New("boooooom")
	// Mock behaviour must be defined before the actual call
	repository.On("Create", mock.Anything,_generateSameTask()).Return(rError)

	err := interactor.Process(context.TODO(), _generateSameTask())

	Expect(err).To(HaveOccurred(), "should return an error")
	Expect(err).To(Equal(rError), "should return same error")
	Expect(repository.AssertExpectations(t)).To(BeTrue(), "all methods where called")
}


func _generateSameTask() models.Task{
	
	return models.NewTask(
		generateSameUlid(),
	    "atask",
		"action",
		interface{}
	)
}

func generateSameUlid() ulid.ULID {
	t := time.Unix(1000000, 0)
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	return ulid.MustNew(ulid.Timestamp(t), entropy)

}
