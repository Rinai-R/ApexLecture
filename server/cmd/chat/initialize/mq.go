package initialize

import (
	"github.com/IBM/sarama"
	"github.com/Rinai-R/ApexLecture/server/cmd/chat/config"
	"github.com/cloudwego/kitex/pkg/klog"
)

func InitMQ() (sarama.AsyncProducer, sarama.ConsumerGroup) {
	ProducerConfig := sarama.NewConfig()
	ProducerConfig.Producer.RequiredAcks = sarama.WaitForAll
	ProducerConfig.Producer.Return.Successes = true
	ProducerConfig.Producer.Partitioner = sarama.NewRandomPartitioner

	producer, err := sarama.NewAsyncProducer(config.GlobalServerConfig.Kafka.Brokers, ProducerConfig)
	if err != nil {
		klog.Fatal("failed to create sync producer: ", err)
	}
	ConsumerConfig := sarama.NewConfig()
	ConsumerConfig.Consumer.Return.Errors = true
	consumer, err := sarama.NewConsumerGroup(config.GlobalServerConfig.Kafka.Brokers, config.GlobalServerConfig.Kafka.Group, ConsumerConfig)
	if err != nil {
		klog.Fatal("failed to create consumer: ", err)
	}
	return producer, consumer
}
