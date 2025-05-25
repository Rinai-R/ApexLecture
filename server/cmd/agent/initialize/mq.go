package initialize

import (
	"time"

	"github.com/IBM/sarama"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/config"
	"github.com/cloudwego/kitex/pkg/klog"
)

func InitMQ() (sarama.AsyncProducer, sarama.ConsumerGroup) {
	// 异步生产者
	ProducerConfig := sarama.NewConfig()
	ProducerConfig.Net.SASL.Enable = true
	ProducerConfig.Net.SASL.User = config.GlobalServerConfig.Kafka.Username
	ProducerConfig.Net.SASL.Password = config.GlobalServerConfig.Kafka.Password
	ProducerConfig.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	ProducerConfig.Net.SASL.Handshake = true
	ProducerConfig.Net.TLS.Enable = false
	ProducerConfig.Net.DialTimeout = 5 * time.Second
	ProducerConfig.Version = sarama.V2_8_0_0

	ProducerConfig.Producer.Return.Successes = true
	ProducerConfig.Producer.RequiredAcks = sarama.WaitForAll
	ProducerConfig.Producer.Idempotent = true
	ProducerConfig.Net.MaxOpenRequests = 1

	producer, err := sarama.NewAsyncProducer(config.GlobalServerConfig.Kafka.Brokers, ProducerConfig)
	if err != nil {
		klog.Fatal("failed to create sync producer: ", err)
	}
	go func() {
		for err := range producer.Errors() {
			klog.Error("Chat MQ producer error: ", err)
		}
	}()

	go func() {
		for msg := range producer.Successes() {
			klog.Infof("Chat MQ producer success, topic: %s, partition: %d", msg.Topic, msg.Partition)
		}
	}()

	// 当前服务启动一个特定的消费者组，为之后异步入库做准备。
	ConsumerConfig := sarama.NewConfig()
	ConsumerConfig.Consumer.Return.Errors = true
	ConsumerConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	ConsumerConfig.Net.SASL.Enable = true
	ConsumerConfig.Net.SASL.User = config.GlobalServerConfig.Kafka.Username
	ConsumerConfig.Net.SASL.Password = config.GlobalServerConfig.Kafka.Password
	ConsumerConfig.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	ConsumerConfig.Net.SASL.Handshake = true
	ConsumerConfig.Net.TLS.Enable = false
	ConsumerConfig.Version = sarama.V2_8_0_0
	consumer, err := sarama.NewConsumerGroup(config.GlobalServerConfig.Kafka.Brokers, config.GlobalServerConfig.Kafka.Group, ConsumerConfig)
	if err != nil {
		klog.Fatal("failed to create consumer: ", err)
	}
	return producer, consumer
}
