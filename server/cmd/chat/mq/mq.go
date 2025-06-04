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
	"github.com/streadway/amqp"
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
	// 对于聊天消息的推送
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

type ConsumerManager interface {
	Consume(ctx context.Context, topic string, handler *ConsumerHandlerImpl) error
}

var _ ConsumerManager = (*ConsumerManagerImpl)(nil)

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
				session.MarkMessage(message, "")
			default:
				klog.Error("Unknown message type", base.InternalMessageType(Message.Type))
			}
		case <-session.Context().Done():
			return nil
		}
	}
}

// ============================= 兔子 mq ========================

type PublisherManagerImpl struct {
	ch       *amqp.Channel
	Exchange string
}

func NewPublisherManager(conn *amqp.Connection, exchange string) *PublisherManagerImpl {
	ch, err := conn.Channel()
	if err != nil {
		klog.Error("NewChannel failed", err)
		return nil
	}
	err = ch.ExchangeDeclare(
		exchange,           // 交换机名称
		amqp.ExchangeTopic, // 交换机类型
		true,               // 是否持久化
		false,              // 是否自动删除
		false,              // 是否排他交换机
		false,              // 是否等待确认（即消息持久化）
		nil,                // 附加参数
	)
	if err != nil {
		klog.Fatal("ExchangeDeclare failed", err)
		return nil
	}
	return &PublisherManagerImpl{ch: ch, Exchange: exchange}
}

func (p *PublisherManagerImpl) SendMessage(ctx context.Context, request *chat.ChatMessage) error {
	var msg *base.InternalMessage
	// 通过公共的消息类型，实现消息结构的统一化。
	switch base.InternalMessageType(request.Type) {
	// 对于聊天消息的推送
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
	err = p.ch.Publish(
		p.Exchange, // 交换机名称
		fmt.Sprintf(consts.RoomKey, request.RoomId), // 路由键
		false, // 是否持久化
		false, // 是否等待确认（即消息持久化）
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        bytes,
		},
	)
	if err != nil {
		klog.Error("Publish failed", err)
		return err
	}
	return nil
}

var _ ConsumerManager = (*SubscriberManagerImpl)(nil)

type SubscriberManagerImpl struct {
	ch       *amqp.Channel
	Exchange string
}

func NewSubscriberManager(conn *amqp.Connection, exchange string) *SubscriberManagerImpl {
	ch, err := conn.Channel()
	if err != nil {
		klog.Error("NewChannel failed", err)
		return nil
	}
	err = ch.ExchangeDeclare(
		exchange,           // 交换机名称
		amqp.ExchangeTopic, // 交换机类型
		true,               // 是否持久化
		false,              // 是否自动删除
		false,              // 是否排他交换机
		false,              // 是否等待确认（即消息持久化）
		nil,                // 附加参数
	)
	if err != nil {
		klog.Fatal("ExchangeDeclare failed", err)
		return nil
	}
	return &SubscriberManagerImpl{ch: ch, Exchange: exchange}
}

func (s *SubscriberManagerImpl) Consume(ctx context.Context, topic string, handler *ConsumerHandlerImpl) error {
	klog.Info("RabbitMQ 消费")
	queue, err := s.ch.QueueDeclare(
		"",    // 队列名称
		false, // 是否持久化
		false, // 是否排他
		false, // 是否自动删除
		false, // 是否阻塞
		nil,   // 附加参数
	)
	if err != nil {
		klog.Error("QueueDeclare failed", err)
		return err
	}
	err = s.ch.QueueBind(
		queue.Name, // 队列名称
		"",
		s.Exchange, // 交换机名称
		false,      // 是否持久化
		nil,        // 附加参数
	)
	if err != nil {
		klog.Error("QueueBind failed", err)
		return err
	}
	msgs, err := s.ch.Consume(
		queue.Name, // 队列名称
		"",         // 用来区分多个消费者的标识符
		false,      // 是否自动应答
		false,      // 是否消费者持久化
		false,      // 是否等待确认（即消息持久化）
		false,      // 是否排他
		nil,        // 附加参数
	)
	if err != nil {
		klog.Error("Consume failed", err)
		return err
	}
	for {
		select {
		case msg := <-msgs:
			var Message base.InternalMessage
			err := sonic.Unmarshal(msg.Body, &Message)
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
				handler.MySQLManager.CreateChatMessage(context.Background(), &model.ChatMessage{
					ID:        id,
					SenderID:  Message.Payload.ChatMessage.UserId,
					RoomID:    Message.Payload.ChatMessage.RoomId,
					Content:   Message.Payload.ChatMessage.Message,
					CreatedAt: time.Now(),
				})
			default:
				klog.Error("Unknown message type", base.InternalMessageType(Message.Type))
			}
			err = msg.Ack(false)
			if err != nil {
				klog.Error("Ack failed", err)
				continue
			}
		case <-ctx.Done():
			return nil
		}
	}
}
