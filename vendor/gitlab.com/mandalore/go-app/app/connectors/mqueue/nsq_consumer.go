package mqueue

import (
	"github.com/nsqio/go-nsq"
	"gitlab.com/mandalore/go-app/app"
	"gitlab.com/vredens/go-logger"
)

// NSQConsumer ...
type NSQConsumer struct {
	conn    *nsq.Consumer
	handler IRequestHandler
	started bool
	config  *NSQConfig
	log     logger.Logger
}

// NewNSQConsumer ...
func NewNSQConsumer(config *NSQConfig) *NSQConsumer {
	cfg := getNSQConsumerConnectorConfig(config)

	log := logger.Spawn(
		logger.WithFields(logger.Fields{"component": "nsq", "topic": config.Topic, "channel": config.Channel}),
		logger.WithTags("consumer"),
	)

	log.Info("creating consumer")

	nsqConsumer, err := nsq.NewConsumer(config.Topic, config.Channel, cfg)
	if err != nil {
		panic(err)
	}

	consumer := &NSQConsumer{
		conn:   nsqConsumer,
		config: config,
		log:    log,
	}

	nsqConsumer.SetLogger(consumer, nsq.LogLevelWarning)

	return consumer
}

// SetLogger changes the consumer's logger to something else.
func (consumer *NSQConsumer) SetLogger(l logger.Logger) {
	if l != nil {
		consumer.log = l
		consumer.conn.SetLogger(consumer, nsq.LogLevelWarning)
	}
}

// SetHandler ...
func (consumer *NSQConsumer) SetHandler(handler IRequestHandler, concurrency uint) error {
	consumer.handler = handler
	consumer.conn.AddConcurrentHandlers(consumer, int(concurrency))

	return nil
}

// Start ...
func (consumer *NSQConsumer) Start() error {
	if consumer.handler == nil {
		return app.NewError(app.ErrorDevPoo, "no handler configured", nil)
	}

	return consumer.start()
}

func (consumer *NSQConsumer) start() error {
	consumer.conn.ChangeMaxInFlight(consumer.config.GetMaxInflight())

	l := consumer.log.Spawn(logger.WithTags("boot"))

	if consumer.config.HasLookupdAddresses() {
		consumer.started = true
		for _, addr := range consumer.config.GetLookupdAddresses() {
			l.Infof("NSQ consumer connecting to %s", addr)
		}
		if err := consumer.conn.ConnectToNSQLookupds(consumer.config.GetLookupdAddresses()); err != nil {
			return err
		}
	}
	if consumer.config.HasNodeAddresses() {
		consumer.started = true
		for _, addr := range consumer.config.GetNodeAddresses() {
			l.Infof("NSQ consumer connecting to %s", addr)
		}
		if err := consumer.conn.ConnectToNSQDs(consumer.config.GetNodeAddresses()); err != nil {
			return err
		}
	}

	if !consumer.started {
		panic("failed to start consumer")
	}

	<-consumer.conn.StopChan

	consumer.started = false

	return nil
}

// IsStarted returns true if the consumer has been started.
func (consumer *NSQConsumer) IsStarted() bool {
	return consumer.started
}

// IsStopped returns true if the consumer has been stopped.
func (consumer *NSQConsumer) IsStopped() bool {
	return !consumer.started
}

// Stop stops the consumer.
func (consumer *NSQConsumer) Stop() error {
	consumer.conn.Stop()

	return nil
}

// HandleMessage ...
func (consumer *NSQConsumer) HandleMessage(nsqMsg *nsq.Message) error {
	msg := newMessage(nsqMsg, consumer.config)

	consumer.log.DebugData("message received", app.KV{"id": msg.GetID(), "data": string(msg.GetContent())})

	err := consumer.handler.HandleRequest(msg)

	if !msg.nsqMsg.HasResponded() {
		if err != nil {
			consumer.log.InfoData("requeing message due to router failure", app.KV{"id": msg.GetID(), "count": msg.GetRetries(), "error": app.StringifyError(err)})
			msg.RequeueWithDelay(consumer.config.GetDefaultRequeueDelay())
		} else {
			msg.Processed()
			consumer.log.DebugData("message processed", app.KV{"id": msg.GetID()})
		}
	}

	return nil
}

// Output implements the nsq.Logger interface Output method
func (consumer *NSQConsumer) Output(level int, msg string) error {
	return nsqOutputLogger(consumer.log, level, msg)
}

// NSQConsumerSimple is a basic consumer which performs only common operations.
type NSQConsumerSimple struct {
	NSQConsumer
	simpleHandler nsq.Handler
}

// NewNSQConsumerSimple creates a new instance of the NSQConsumerSimple.
func NewNSQConsumerSimple(config *NSQConfig) *NSQConsumerSimple {
	consumer := NewNSQConsumer(config)

	return &NSQConsumerSimple{
		NSQConsumer: *consumer,
	}
}

// SetHandler ...
func (consumer *NSQConsumerSimple) SetHandler(handler nsq.Handler, concurrency uint) error {
	consumer.simpleHandler = handler
	consumer.conn.AddConcurrentHandlers(consumer, int(concurrency))

	return nil
}

// Start ...
func (consumer *NSQConsumerSimple) Start() error {
	if consumer.simpleHandler == nil {
		return app.NewError(app.ErrorDevPoo, "no handler configured", nil)
	}

	return consumer.start()
}

// HandleMessage required for integration with NSQ Consumer.
func (consumer *NSQConsumerSimple) HandleMessage(msg *nsq.Message) error {
	consumer.log.DebugData("message received", app.KV{"id": msg.ID, "data": string(msg.Body)})

	err := consumer.simpleHandler.HandleMessage(msg)
	if err != nil {
		if !msg.HasResponded() {
			consumer.log.InfoData("requeing message due to router failure", app.KV{"id": msg.ID, "count": msg.Attempts, "error": app.StringifyError(err)})
			msg.RequeueWithoutBackoff(consumer.config.GetDefaultRequeueDelay())
		}
		return err
	}

	if !msg.HasResponded() {
		msg.Finish()
		consumer.log.DebugData("message processed", app.KV{"id": msg.ID, "data": string(msg.Body)})
	}

	return nil
}
