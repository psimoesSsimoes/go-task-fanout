package mqueue

import (
	"time"

	nsq "github.com/nsqio/go-nsq"

	"gitlab.com/vredens/go-logger"
)

// IRequestHandler ...
type IRequestHandler interface {
	HandleRequest(IMessage) error
}

// IMessage represents the internal message
type IMessage interface {
	// GetID returns the unique identifier of this message in the underlying messaging system.
	GetID() string
	// GetContent returns the message payload in byte array format.
	GetContent() []byte
	// GetContentType returns the message's content type.
	GetContentType() string
	// GetRetries returns the number of times the message has been requeued (always 0 if unsupported by the protocol).
	GetRetries() uint
	// Requeue sends the message back to the queue to be processed after the default delay.
	Requeue() error
	// RequeueWithDelay is the same as requeue except it allows the delay to be specified.
	RequeueWithDelay(delay time.Duration) error
	// Processed marks the message as processed and removes it from the queue.
	Processed() error
}

// NSQConfig ...
type NSQConfig struct {
	Lookupd      []string `mapstructure:"lookupd"`
	Nsqd         []string `mapstructure:"nsqd"`
	Topic        string   `mapstructure:"topic"`
	Channel      string   `mapstructure:"channel"`
	RequeueDelay int64    `mapstructure:"requeue_delay"`
	MaxInflight  int      `mapstructure:"max_inflight"`
	MaxAttempts  uint16   `mapstructure:"max_attempts"`
}

// NewConfig ...
func NewConfig(topic, channel string, lookupd, nsqd []string) *NSQConfig {
	cfg := &NSQConfig{
		Topic:        topic,
		Channel:      channel,
		Lookupd:      lookupd,
		Nsqd:         nsqd,
		RequeueDelay: 30,
		MaxInflight:  5,
		MaxAttempts:  24 * 120, // one day of attempts, assuming default requeue delay of 30 seconds
	}

	return cfg
}

// SetRequeueDelay ...
func (config *NSQConfig) SetRequeueDelay(delay time.Duration) {
	config.RequeueDelay = int64(delay / time.Second)
}

// SetMaxInflight ...
func (config *NSQConfig) SetMaxInflight(size int) {
	config.MaxInflight = size
}

// GetChannel ...
func (config *NSQConfig) GetChannel() string {
	return config.Channel
}

// GetTopic ...
func (config *NSQConfig) GetTopic() string {
	return config.Topic
}

// HasLookupdAddresses returns true if there are lookupd addresses configured.
func (config *NSQConfig) HasLookupdAddresses() bool {
	return config.Lookupd != nil && len(config.Lookupd) > 0
}

// GetLookupdAddresses ...
func (config *NSQConfig) GetLookupdAddresses() []string {
	return config.Lookupd
}

// HasNodeAddresses returns true if there are NSQDaemon addresses configured.
func (config *NSQConfig) HasNodeAddresses() bool {
	return config.Nsqd != nil && len(config.Nsqd) > 0
}

// GetNodeAddresses ...
func (config *NSQConfig) GetNodeAddresses() []string {
	return config.Nsqd
}

// GetDefaultRequeueDelay ...
func (config *NSQConfig) GetDefaultRequeueDelay() time.Duration {
	return time.Duration(config.RequeueDelay) * time.Second
}

// GetMaxInflight ...
func (config *NSQConfig) GetMaxInflight() int {
	return config.MaxInflight
}

// GetProcessTimeout sets the timeout for the application to process a message.
func (config *NSQConfig) GetProcessTimeout() time.Duration {
	return 120 * time.Second
}

func nsqOutputLogger(log logger.Logger, level int, msg string) error {
	switch level {
	case int(nsq.LogLevelDebug):
		log.Debug(msg)
	case int(nsq.LogLevelInfo):
		log.Debug(msg)
	case int(nsq.LogLevelWarning):
		log.Warn(msg)
	case int(nsq.LogLevelError):
		log.Error(msg)
	}

	return nil
}

func getNSQConsumerConnectorConfig(config *NSQConfig) *nsq.Config {
	if config.Topic == "" {
		// TODO: error
		panic("bad config")
	}
	if config.Channel == "" {
		// TODO: error
		panic("bad config")
	}
	if len(config.Lookupd) == 0 && len(config.Nsqd) == 0 {
		// TODO: error
		panic("bad config")
	}
	if config.RequeueDelay == 0 {
		config.RequeueDelay = 30
	}
	if config.MaxInflight == 0 {
		config.MaxInflight = 5
	}

	return generateNSQConnectorConfig(config)
}

func getNSQProducerConnectorConfig(config *NSQConfig) *nsq.Config {
	if config.Topic == "" {
		// TODO: error
		panic("bad config")
	}
	if len(config.Lookupd) == 0 && len(config.Nsqd) == 0 {
		// TODO: error
		panic("bad config")
	}

	return generateNSQConnectorConfig(config)
}

func generateNSQConnectorConfig(config *NSQConfig) *nsq.Config {
	cfg := nsq.NewConfig()

	cfg.MaxAttempts = 0
	if config.MaxAttempts > 0 {
		cfg.MaxAttempts = config.MaxAttempts
	}
	if config.GetDefaultRequeueDelay() > 0 {
		cfg.DefaultRequeueDelay = config.GetDefaultRequeueDelay()
	}
	if config.GetMaxInflight() > 0 {
		cfg.MaxInFlight = config.GetMaxInflight()
	}
	if config.GetProcessTimeout() > 0 {
		cfg.ReadTimeout = config.GetProcessTimeout()
	}

	return cfg
}
