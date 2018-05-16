package mqueue

import (
	"github.com/nsqio/go-nsq"
	logger "gitlab.com/vredens/go-logger"
)

// NSQProducer ...
type NSQProducer struct {
	conn   *nsq.Producer
	config *NSQConfig
	log    logger.Logger
}

// NewNSQProducer ...
func NewNSQProducer(config *NSQConfig) *NSQProducer {
	var addr string
	cfg := getNSQProducerConnectorConfig(config)

	if config.HasLookupdAddresses() {
		addresses := config.GetLookupdAddresses()

		// TODO: get a random address from the list instead
		addr = addresses[0]
	} else if config.HasNodeAddresses() {
		addresses := config.GetNodeAddresses()

		// TODO: get a random address from the list instead
		addr = addresses[0]
	} else {
		panic("no NSQ address to connect to")
	}

	nsqProducer, err := nsq.NewProducer(addr, cfg)
	if err != nil {
		panic(err)
	}

	producer := &NSQProducer{
		conn:   nsqProducer,
		config: config,
		log: logger.Spawn(
			logger.WithFields(logger.Fields{"component": "nsq"}),
			logger.WithTags("publish"),
		),
	}

	producer.conn.SetLogger(producer, nsq.LogLevelWarning)

	return producer
}

// Publish published a message on to a topic.
func (producer *NSQProducer) Publish(topic string, body []byte, maxRetries int) error {
	var err error

	for count := 0; count < maxRetries; count++ {
		if err = producer.conn.Publish(topic, body); err == nil {
			return nil
		}
	}

	return err
}

// Output used to log output for the underlying nsq producer.
func (producer *NSQProducer) Output(level int, msg string) error {
	return nsqOutputLogger(producer.log, level, msg)
}

// Stop ...
func (producer *NSQProducer) Stop() {
	producer.conn.Stop()
}

// Ping ...
func (producer *NSQProducer) Ping() error {
	return producer.conn.Ping()
}
