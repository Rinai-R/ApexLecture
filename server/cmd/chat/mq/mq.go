package mq

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/chat"
)

type ProducerManager struct {
	producer sarama.SyncProducer
}

func NewProducerManager(producer sarama.SyncProducer) *ProducerManager {
	return &ProducerManager{producer: producer}
}

func (p *ProducerManager) SendMessage(ctx context.Context, request *chat.ChatMessage) (err error) {
	return nil
}

type ConsumerManager struct {
	consumer sarama.Consumer
}

func NewConsumerManager(consumer sarama.Consumer) *ConsumerManager {
	return &ConsumerManager{consumer: consumer}
}

func (c *ConsumerManager) Consume(ctx context.Context, topic string) (err error) {
	return nil
}
