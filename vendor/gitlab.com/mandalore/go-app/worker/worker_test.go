package worker

import (
	"fmt"
	"io/ioutil"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/onsi/gomega" // testing assertions
	"gitlab.com/mandalore/go-app/app"
	"gitlab.com/vredens/go-logger"
)

var mute = logger.NewWriter(logger.WithLevel(logger.OFF), logger.WithOutput(ioutil.Discard)).Spawn()

func newWorkingHandler(id string, data interface{}, control chan bool) WorkHandler {
	return func(tid string, tdata interface{}) error {
		if tid != id {
			panic("failed!")
		}
		if tdata != data {
			panic("data failed")
		}
		// Expect(tid).To(Equal(id))
		// Expect(tdata).To(Equal(tdata))

		control <- true

		return nil
	}
}

func TestWorker(t *testing.T) {
	var err error
	ctrl := make(chan bool)
	doneChan := make(chan bool)
	w := NewWorker(WithLogger(mute), WithWorkHandler(newWorkingHandler("task", "a name", ctrl)))

	if err = w.Process("task", "a name"); err != nil {
		t.Error("expected worker Process to not return an error")
	}

	go func() {
		if err := w.Start(); err != nil {
			t.Error("expected worker Start to not return an error")
		}
		doneChan <- true
	}()

	select {
	case <-ctrl:
		t.Log("OK")
	case <-time.After(1100 * time.Millisecond):
		t.Error("failed to process task in time")
	}

	if err = w.Stop(); err != nil {
		t.Error("expected worker stop to not return an error")
	}
	<-doneChan
}

func TestWorkerTimeout(t *testing.T) {
	RegisterTestingT(t)

	var err error
	ctrl := make(chan bool)
	doneChan := make(chan bool)
	w := NewWorker(WithLogger(mute), WithWorkHandler(newWorkingHandler("task", "a name", ctrl)))

	if err = w.Process("task", "a name"); err != nil {
		t.Error("expecte worker Process to not return an error")
	}

	go func() {
		if err := w.Start(); err != nil {
			t.Error("expected worker Start to not return an error")
		}
		doneChan <- true
	}()

	select {
	case <-ctrl:
		t.Log("OK")
	case <-time.After(1100 * time.Millisecond):
		t.Error("failed to process task in time")
	}

	if err = w.Stop(); err != nil {
		t.Error("expected worker stop return an error")
	}
	<-doneChan
}

func TestWorkerProcessingDelay(t *testing.T) {
	RegisterTestingT(t)

	ctrl := make(chan bool)
	var incr uint32
	var err error
	w := NewWorker(
		WithLogger(mute),
		WithWorkHandler(func(tid string, data interface{}) error {
			atomic.AddUint32(&incr, 1)
			return nil
		}),
		WithMaxPerSecond(10),
		WithProcessDelay(110),
		WithTick(100),
	)

	var i uint32
	for i = 0; i < 2; i++ {
		err := w.Process(fmt.Sprintf("%d", i), i)
		Expect(err).To(BeNil())
	}

	go func() {
		ctrl <- true
		err := w.Start()
		Expect(err).To(BeNil())
		ctrl <- true
	}()

	<-ctrl
	for i = 0; i <= 2; i++ {
		<-time.After(105 * time.Millisecond)
		val := atomic.LoadUint32(&incr)
		Expect(val).To(Equal(uint32(i)))
	}

	err = w.Stop()
	Expect(err).To(BeNil())
	<-ctrl
	Expect(w.queue.Size()).To(Equal(0))
}

func TestWorkerBurstLimit(t *testing.T) {
	RegisterTestingT(t)

	var incr uint32

	var err error
	w := NewWorker(
		WithLogger(mute),
		WithWorkHandler(func(tid string, data interface{}) error {
			val := data.(uint32)
			atomic.AddUint32(&incr, val)
			return nil
		}),
		WithMaxPerSecond(10),
		WithProcessDelay(0),
		WithTick(100),
	)

	var i uint32
	for i = 0; i < 11; i++ {
		err := w.Process(fmt.Sprintf("%d", i), i)
		Expect(err).To(BeNil())
	}

	ctrlChan := make(chan bool)

	go func() {
		ctrlChan <- true
		err := w.Start()
		Expect(err).To(BeNil())
		ctrlChan <- true
	}()

	<-ctrlChan
	<-time.After(1050 * time.Millisecond)
	val := atomic.LoadUint32(&incr)
	Expect(val).To(Equal(uint32(45)))

	err = w.Kill()
	Expect(err).To(BeNil())

	<-ctrlChan
	Expect(w.queue.Size()).To(Equal(1))
}

func TestWorkerSoftShutdown(t *testing.T) {
	RegisterTestingT(t)

	var incr uint32

	var err error
	w := NewWorker(
		WithLogger(mute),
		WithWorkHandler(func(tid string, data interface{}) error {
			val := data.(uint32)
			atomic.AddUint32(&incr, val)
			<-time.After(50 * time.Millisecond)
			return nil
		}),
		WithMaxPerSecond(10),
		WithProcessDelay(0),
		WithTick(100),
	)

	doneChan := make(chan bool)

	var i uint32
	for i = 0; i < 11; i++ {
		err := w.Process(fmt.Sprintf("%d", i), i)
		Expect(err).To(BeNil())
	}

	go func() {
		err := w.Start()
		Expect(err).To(BeNil())
		doneChan <- true
	}()

	<-time.After(100 * time.Millisecond)

	err = w.Stop()
	Expect(err).To(BeNil())

	i++
	err = w.Process(fmt.Sprintf("%d", i), i)
	Expect(err).ToNot(BeNil())
	aerr := err.(app.Error)
	Expect(aerr.GetCode()).To(Equal(app.ErrorConflict))

	select {
	case <-doneChan:
		val := atomic.LoadUint32(&incr)
		Expect(val).To(Equal(uint32(55)))
		Expect(w.queue.Size()).To(Equal(0))
	case <-time.After(5000 * time.Millisecond):
		t.Fail()
	}
}

func TestWorkerHardShutdown(t *testing.T) {
	RegisterTestingT(t)

	var incr uint32

	var err error
	w := NewWorker(
		WithLogger(mute),
		WithWorkHandler(func(tid string, data interface{}) error {
			val := data.(uint32)
			atomic.AddUint32(&incr, val)
			return nil
		}),
		WithMaxPerSecond(10),
		WithProcessDelay(0),
		WithTick(100),
	)

	ctrlChan := make(chan bool)

	var i uint32
	for i = 0; i < 11; i++ {
		err := w.Process(fmt.Sprintf("%d", i), i)
		Expect(err).To(BeNil())
	}

	go func() {
		ctrlChan <- true
		err := w.Start()
		Expect(err).To(BeNil())
		ctrlChan <- true
	}()

	<-ctrlChan
	// allow time to process 2 ticks
	<-time.After(220 * time.Millisecond)
	err = w.Kill()
	Expect(err).To(BeNil())

	i++
	err = w.Process(fmt.Sprintf("%d", i), i)
	Expect(err).ToNot(BeNil())
	aerr := err.(app.Error)
	Expect(aerr.GetCode()).To(Equal(app.ErrorConflict))

	select {
	case <-ctrlChan:
		val := atomic.LoadUint32(&incr)
		Expect(val).To(Equal(uint32(1)))
		Expect(w.queue.Size()).To(Equal(9))
	case <-time.After(5000 * time.Millisecond):
		t.Fail()
	}
}
