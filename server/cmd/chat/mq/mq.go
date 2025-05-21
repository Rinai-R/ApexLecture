package mq

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/Rinai-R/ApexLecture/server/cmd/chat/config"
	"github.com/Rinai-R/ApexLecture/server/cmd/chat/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/chat/model"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/chat"
	"github.com/bwmarrin/snowflake"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/kitex/pkg/klog"
)

type ProducerManagerImpl struct {
	producer sarama.AsyncProducer
}

func NewProducerManager(producer sarama.AsyncProducer) *ProducerManagerImpl {
	return &ProducerManagerImpl{producer: producer}
}

func (p *ProducerManagerImpl) SendMessage(ctx context.Context, request *chat.ChatMessage) error {
	bytes, err := sonic.Marshal(request)
	if err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{
		Topic: config.GlobalServerConfig.Kafka.Topic,
		Key:   sarama.StringEncoder(fmt.Sprintf("room:%d", request.RoomId)),
		Value: sarama.StringEncoder(string(bytes)),
	}
	p.producer.Input() <- msg
	return nil
}

type ConsumerManagerImpl struct {
	consumer sarama.ConsumerGroup
}

func NewConsumerManager(consumer sarama.ConsumerGroup) *ConsumerManagerImpl {
	return &ConsumerManagerImpl{consumer: consumer}
}

func (c *ConsumerManagerImpl) Consume(ctx context.Context, topic string, handler *ConsumerHandlerImpl) (err error) {
	for {
		err = c.consumer.Consume(ctx, []string{topic}, handler)
		if err != nil {
			klog.Error("Consume failed", err)
		}
		if ctx.Err() != nil {
			klog.Info("Kafka 消费退出")
			break
		}
	}
	return
}

// ============================= 下面是关于 Consumer Handler 的部分 =============================

type ConsumerHandlerImpl struct {
	LoveFrom5mm  string
	MySQLManager *dao.MysqlManagerImpl
}

var _ sarama.ConsumerGroupHandler = (*ConsumerHandlerImpl)(nil)

func NewConsumerHandler(mysqlManager *dao.MysqlManagerImpl) *ConsumerHandlerImpl {
	return &ConsumerHandlerImpl{
		LoveFrom5mm:  "曹寺5mm",
		MySQLManager: mysqlManager,
	}
}

func (h *ConsumerHandlerImpl) Setup(session sarama.ConsumerGroupSession) error {
	klog.Info("ConsumerHandlerImpl: Setup")
	return nil
}

func (h *ConsumerHandlerImpl) Cleanup(session sarama.ConsumerGroupSession) error {
	klog.Info("ConsumerHandlerImpl: Cleanup")
	return nil
}

func (h *ConsumerHandlerImpl) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	klog.Info("ConsumerHandlerImpl: ConsumeClaim")
	for message := range claim.Messages() {
		var chatMessage chat.ChatMessage
		err := sonic.Unmarshal(message.Value, &chatMessage)
		if err != nil {
			klog.Error("Unmarshal failed", err)
			continue
		}
		sf, err := snowflake.NewNode(consts.MessageIDSnowFlakeNode)
		if err != nil {
			klog.Error("NewNode failed", err)
			continue
		}
		id := sf.Generate().Int64()
		h.MySQLManager.CreateChatMessage(context.Background(), &model.ChatMessage{
			ID:        id,
			SenderID:  chatMessage.UserId,
			RoomID:    chatMessage.RoomId,
			Content:   chatMessage.Text,
			CreatedAt: time.Now(),
		})
	}
	return nil
}
