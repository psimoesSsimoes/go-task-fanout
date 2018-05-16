package models

import (
	"math/rand"
	"testing"
	"time"

	"github.com/oklog/ulid"
	. "github.com/onsi/gomega"
)

func TestCreateTask(t *testing.T) {
	RegisterTestingT(t)
	var (
		i     interface{}
		atime = time.Date(2019, time.January, 27, 10, 23, 0, 0, time.UTC)
	)

	task := NewTask(generateSameUlid(),
		"taskid",
		"action",
		i,
		WithCreatedAt(atime),
		WithStartedAt(atime))

	Expect(task.ID).To(Equal(generateSameUlid()), "should have set the correct ulid")
	Expect(task.TaskID).To(Equal("taskid"), "should have set taskid")
	Expect(task.Action).To(Equal("action"), "should have set action")
	Expect(task.Data).To(BeNil(), "should have set data")
	Expect(task.CreatedAt.Format(time.RFC822)).To(Equal(atime.Format(time.RFC822)), "should have set CreatedAt")
	Expect(task.StartedAt.Format(time.RFC822)).To(Equal(atime.Format(time.RFC822)), "should have set StartedAt")

}

func generateSameUlid() ulid.ULID {
	t := time.Unix(1000000, 0)
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	return ulid.MustNew(ulid.Timestamp(t), entropy)

}
