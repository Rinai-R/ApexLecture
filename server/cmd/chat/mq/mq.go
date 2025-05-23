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
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/base"
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
	var msg *base.InternalMessage
	// 通过公共的消息类型，实现消息结构的统一化。
	switch base.InternalMessageType(request.Type) {
	case base.InternalMessageType_CHAT_MESSAGE:
		msg = &base.InternalMessage{
			Type: base.InternalMessageType_CHAT_MESSAGE,
			Payload: &base.InternalPayload{
				ChatMessage: &base.InternalChatMessage{
					Message: request.Text,
					RoomId:  request.RoomId,
					UserId:  request.UserId,
				},
			},
		}
	}
	bytes, err := sonic.Marshal(msg)
	if err != nil {
		return err
	}
	p.producer.Input() <- &sarama.ProducerMessage{
		Topic: config.GlobalServerConfig.Kafka.Topic,
		Key:   sarama.StringEncoder(fmt.Sprintf(consts.RoomKey, request.RoomId)),
		Value: sarama.StringEncoder(string(bytes)),
	}
	return nil
}

type ConsumerManagerImpl struct {
	consumer sarama.ConsumerGroup
}

func NewConsumerManager(consumer sarama.ConsumerGroup) *ConsumerManagerImpl {
	return &ConsumerManagerImpl{consumer: consumer}
}

func (c *ConsumerManagerImpl) Consume(ctx context.Context, topic string, handler *ConsumerHandlerImpl) error {
	klog.Info("Kafka 消费")
	err := c.consumer.Consume(ctx, []string{topic}, handler)
	if err != nil {
		klog.Error("Consume failed", err)
		return err
	}
	if ctx.Err() != nil {
		klog.Info("Kafka 消费退出")
		return ctx.Err()
	}

	return nil
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
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				klog.Info("ConsumerHandlerImpl: 消费完毕")
				return nil
			}
			var Message base.InternalMessage
			err := sonic.Unmarshal(message.Value, &Message)
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
			// 识别字段类型，确保正常处理。
			switch base.InternalMessageType(Message.Type) {
			case base.InternalMessageType_CHAT_MESSAGE:
				h.MySQLManager.CreateChatMessage(context.Background(), &model.ChatMessage{
					ID:        id,
					SenderID:  Message.Payload.ChatMessage.UserId,
					RoomID:    Message.Payload.ChatMessage.RoomId,
					Content:   Message.Payload.ChatMessage.Message,
					CreatedAt: time.Now(),
				})
			default:
				klog.Error("Unknown message type", base.InternalMessageType(Message.Type))
			}
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}
