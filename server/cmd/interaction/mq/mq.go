package mq

import (
	"context"

	"github.com/IBM/sarama"
)

type ProducerManager struct {
	producer sarama.SyncProducer
}

type ConsumerManager struct {
	consumer sarama.Consumer
}

func NewProducerManager(producer sarama.SyncProducer) *ProducerManager {
	return &ProducerManager{producer: producer}
}

func (m *ProducerManager) SendMessage(ctx context.Context, message *sarama.ProducerMessage) error {
	return nil
}

func NewConsumerManager(consumer sarama.Consumer) *ConsumerManager {
	return &ConsumerManager{consumer: consumer}
}

func (m *ConsumerManager) Consume(ctx context.Context, topic string) error {
	return nil
}
