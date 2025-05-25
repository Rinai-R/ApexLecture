package mq

import (
	"context"
	"fmt"
	"strconv"

	"github.com/IBM/sarama"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/config"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/dao"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/cloudwego/kitex/pkg/klog"
)

type ProducerManagerImpl struct {
	producer sarama.AsyncProducer
}

func NewProducerManager(producer sarama.AsyncProducer) *ProducerManagerImpl {
	return &ProducerManagerImpl{
		producer: producer,
	}
}

func (p *ProducerManagerImpl) Send(ctx context.Context, RoomId int64) error {
	p.producer.Input() <- &sarama.ProducerMessage{
		Topic: config.GlobalServerConfig.Kafka.Topic,
		Key:   sarama.StringEncoder(fmt.Sprintf(consts.RoomKey, RoomId)),
		Value: sarama.StringEncoder(fmt.Sprintf("%d", RoomId)),
	}
	return nil
}

type ConsumerManager interface {
}

type ConsumerManagerImpl struct {
	consumer sarama.ConsumerGroup
}

var _ ConsumerManager = (*ConsumerManagerImpl)(nil)

func NewConsumerManager(consumer sarama.ConsumerGroup) *ConsumerManagerImpl {
	return &ConsumerManagerImpl{
		consumer: consumer,
	}
}

func (c *ConsumerManagerImpl) Consume(ctx context.Context, topic string, handler *ConsumerHandlerImpl) error {
	klog.Info("Kafka 消费")
	err := c.consumer.Consume(ctx, []string{topic}, handler)
	if err != nil {
		klog.Error("Consume failed", err)
		return err
	}

	return nil
}

type ConsumerHandlerImpl struct {
	MysqlManager *dao.MysqlManagerImpl
}

func NewConsumerHandler(MysqlManager *dao.MysqlManagerImpl) *ConsumerHandlerImpl {
	return &ConsumerHandlerImpl{
		MysqlManager: MysqlManager,
	}
}

func (h *ConsumerHandlerImpl) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerHandlerImpl) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerHandlerImpl) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				klog.Info("Kafka 消费退出")
			}
			RoomId, err := strconv.ParseInt(string(msg.Value), 10, 64)
			if err != nil {
				klog.Error("KafkaConsumer:ParseInt failed", err)
				continue
			}
			klog.Info("Kafka 消费", RoomId)
			// 这里需要从 Minio 获取 ogg 文件，通过第三方包转换成文本，
			// 然后交给 agent 分析并且存储到 MySQL 数据库中
			// 暂时没找到可以转换的对应的第三方包
			// TODO
		case <-session.Context().Done():
			return nil
		}
	}
}
