package initialize

import (
	"github.com/IBM/sarama"
	"github.com/Rinai-R/ApexLecture/server/cmd/interaction/config"
	"github.com/cloudwego/kitex/pkg/klog"
)

func InitMQ() (sarama.SyncProducer, sarama.Consumer) {
	ProducerConfig := sarama.NewConfig()
	ProducerConfig.Producer.RequiredAcks = sarama.WaitForAll
	ProducerConfig.Producer.Return.Successes = true
	ProducerConfig.Producer.Partitioner = sarama.NewRandomPartitioner

	producer, err := sarama.NewSyncProducer(config.GlobalServerConfig.Kafka.Brokers, ProducerConfig)
	if err != nil {
		klog.Fatal("failed to create sync producer: ", err)
	}
	ConsumerConfig := sarama.NewConfig()
	ConsumerConfig.Consumer.Return.Errors = true
	consumer, err := sarama.NewConsumer(config.GlobalServerConfig.Kafka.Brokers, ConsumerConfig)
	if err != nil {
		klog.Fatal("failed to create consumer: ", err)
	}
	return producer, consumer
}
