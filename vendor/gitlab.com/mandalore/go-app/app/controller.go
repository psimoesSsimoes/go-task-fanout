// Copyright © 2017 JB Ribeiro <self@vredens.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package app aims at providing very simple common application constructs and helpers.
package app

import (
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	log "gitlab.com/vredens/go-logger"
)

// Process is the basic interface for any assynchronous process launched and stopped by the process manager.
type Process interface {
	// Start should block while running the processes and right before terminating should write to the control chan.
	Start() error
	Stop() error
}

const (
	ProcessStateInitialized int32 = iota
	ProcessStateStarted
	ProcessStateTerminated
	ProcessStateAborted
)

type processController struct {
	process Process
	control chan int
	state   int32
}

func (controller processController) Start() {
	controller.process.Start()
}

// ProcessManager handles your processes.
type ProcessManager struct {
	processes map[string]*processController
	control   chan int
	Logger    log.Logger
	started   bool
	mux       sync.RWMutex
}

// NewProcessManager creates a new instance of a ProcessManager.
func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		control:   make(chan int),
		processes: make(map[string]*processController),
		Logger:    log.Spawn(log.WithFields(log.Fields{"component": "pman"})),
	}
}

// IsStarted returns true if the ProcessManager has already started.
func (manager *ProcessManager) IsStarted() bool {
	manager.mux.RLock()
	if manager.started {
		manager.mux.RUnlock()
		return true
	}
	manager.mux.RUnlock()
	return false
}

func (manager *ProcessManager) setStartedTo(isStarted bool) {
	manager.mux.Lock()
	defer manager.mux.Unlock()
	manager.started = isStarted
}

func (manager *ProcessManager) checkAndSetStarted(check bool, set bool) bool {
	manager.mux.Lock()
	defer manager.mux.Unlock()
	if manager.started != check {
		return false
	}
	manager.started = set
	return true
}

// AddProcess stores a proces in the list of processes controlled by the ProcessManager.
func (manager *ProcessManager) AddProcess(name string, process Process) {
	if manager.IsStarted() {
		panic("can not add processes after start")
	}

	manager.processes[name] = &processController{
		process: process,
		control: make(chan int),
		state:   ProcessStateInitialized,
	}
}

func (manager *ProcessManager) launch(name string, pController *processController) {
	atomic.StoreInt32(&pController.state, ProcessStateStarted)
	if err := pController.process.Start(); err != nil {
		manager.Stop()
		manager.Logger.Errorf("process terminating [process:%s]; %s", name, StringifyError(err))
		atomic.StoreInt32(&pController.state, ProcessStateAborted)
		pController.control <- 0
	} else {
		atomic.StoreInt32(&pController.state, ProcessStateTerminated)
		pController.control <- 0
	}
}

// Start blocks until it receives a signal in its control channel or a SIGTERM,
// SIGINT or SIGUSR1, and should be the last method in your main.
func (manager *ProcessManager) Start() {
	if len(manager.processes) < 1 {
		manager.Logger.Errorf("no processes are registered")
		return
	}

	manager.setStartedTo(true)

	// listen for termination signals
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

	manager.mux.Lock()
	for name, process := range manager.processes {
		manager.Logger.Infof("starting [process:%s]", name)
		go manager.launch(name, process)
		manager.Logger.Debugf("started [process:%s]", name)
	}
	manager.mux.Unlock()

	select {
	case <-termChan:
		manager.Logger.Infof("received term signal")
	case <-manager.control:
		manager.Logger.Infof("received shutdown signal")
	}

	manager.setStartedTo(false)

	for name, process := range manager.processes {
		manager.Logger.Infof("stopping process [process:%s]", name)
		if err := process.process.Stop(); err != nil {
			manager.Logger.Errorf("error stopping process [process:%s]; %s", name, StringifyError(err))
		} else {
			manager.Logger.Infof("waiting for process to terminate [process:%s]", name)
			<-process.control
			manager.Logger.Infof("stopped process [process:%s]", name)
		}
	}
}

// Stop will signal the ProcessManager to stop.
func (manager *ProcessManager) Stop() {
	if manager.checkAndSetStarted(true, false) {
		manager.Logger.Debugf("stopping the process manager")
		manager.control <- 1
		manager.Logger.Debugf("process manager stopped")
	}
}

// Destroy removes all processes and closes all channels.
func (manager *ProcessManager) Destroy() map[string]int32 {
	if manager.IsStarted() {
		manager.Stop()
	}
	out := map[string]int32{}
	for name, process := range manager.processes {
		out[name] = atomic.LoadInt32(&process.state)
		close(process.control)
		delete(manager.processes, name)
	}

	return out
}

// StatusCheck returns a tupple where the first value is a bool indicating if all processes are OK, second value is a map for de individual status of each process.
func (manager *ProcessManager) StatusCheck() (bool, map[string]int32) {
	statuses := map[string]int32{}
	status := true

	for n, p := range manager.processes {
		statuses[n] = atomic.LoadInt32(&p.state)
		if atomic.LoadInt32(&p.state) == ProcessStateAborted {
			status = false
		}
	}

	return status, statuses
}
