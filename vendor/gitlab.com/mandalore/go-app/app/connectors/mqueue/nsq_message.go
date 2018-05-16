package mqueue

import (
	"fmt"
	"time"

	nsqlib "github.com/nsqio/go-nsq"
)

type message struct {
	id     string
	topic  string
	addr   string
	config *NSQConfig
	nsqMsg *nsqlib.Message
}

func newMessage(rawMsg *nsqlib.Message, config *NSQConfig) *message {
	msg := &message{
		config: config,
		addr:   rawMsg.NSQDAddress,
		nsqMsg: rawMsg,
	}

	if rawMsg != nil {
		msg.id = fmt.Sprintf("%s", rawMsg.ID)
	}

	return msg
}

// GetID ...
func (msg *message) GetID() string {
	return msg.id
}

// GetContent ...
func (msg *message) GetContent() []byte {
	if msg.nsqMsg == nil {
		return nil
	}

	return msg.nsqMsg.Body
}

// GetContentType ...
func (msg *message) GetContentType() string {
	return msg.config.GetTopic()
}

// GetRetries ...
func (msg *message) GetRetries() uint {
	return uint(msg.nsqMsg.Attempts)
}

func (msg *message) Processed() error {
	if msg.nsqMsg.HasResponded() {
		return fmt.Errorf("message has already responded")
	}

	msg.nsqMsg.Finish()

	return nil
}

func (msg *message) Requeue() error {
	if msg.nsqMsg.HasResponded() {
		return fmt.Errorf("message has already responded")
	}

	msg.nsqMsg.RequeueWithoutBackoff(msg.config.GetDefaultRequeueDelay())

	return nil
}

func (msg *message) RequeueWithDelay(delay time.Duration) error {
	if msg.nsqMsg.HasResponded() {
		return fmt.Errorf("message has already responded")
	}

	msg.nsqMsg.RequeueWithoutBackoff(delay)

	return nil
}
