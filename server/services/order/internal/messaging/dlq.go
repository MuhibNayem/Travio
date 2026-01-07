package messaging

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

type DLQProducer interface {
	PublishError(sagaID, stepName, failureReason string, payload interface{}) error
}

type KafkaDLQProducer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewKafkaDLQProducer(brokers []string, topic string) (*KafkaDLQProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &KafkaDLQProducer{
		producer: producer,
		topic:    topic,
	}, nil
}

type DLQMessage struct {
	SagaID        string      `json:"saga_id"`
	StepName      string      `json:"step_name"`
	FailureReason string      `json:"failure_reason"`
	Payload       interface{} `json:"payload"`
	Timestamp     time.Time   `json:"timestamp"`
}

func (p *KafkaDLQProducer) PublishError(sagaID, stepName, failureReason string, payload interface{}) error {
	msg := DLQMessage{
		SagaID:        sagaID,
		StepName:      stepName,
		FailureReason: failureReason,
		Payload:       payload,
		Timestamp:     time.Now(),
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal dlq message: %w", err)
	}

	kafkaMsg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(sagaID),
		Value: sarama.ByteEncoder(bytes),
	}

	_, _, err = p.producer.SendMessage(kafkaMsg)
	return err
}
