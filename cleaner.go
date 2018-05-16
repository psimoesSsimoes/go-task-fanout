package taskworker

// Cleaner is a task cleaner to a specific command
type Cleaner struct {
	stopChan chan bool
	storage  TaskStorage
}

// NewCleaner creates a new cleaner
func NewCleaner(storage TaskStorage) *Cleaner {
	return &Cleaner{
		stopChan: make(chan bool),
		storage:  storage,
	}
}

// Start starts the process
func (c *Cleaner) Start() error {
	<-c.stopChan

	return nil
}

// Stop stops the process
func (c *Cleaner) Stop() error {
	close(c.stopChan)

	return nil
}

// Kill kills the process
func (c *Cleaner) Kill() error {
	close(c.stopChan)

	return nil
}
