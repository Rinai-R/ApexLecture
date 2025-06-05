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

// 前面是使用 kafka 的部分，已经弃用
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
	klog.Info("RabbitMQ 发布")
	err = p.ch.Publish(
		p.Exchange, // 交换机名称
		"",
		false, // 是否持久化
		false, // 是否等待确认（即消息持久化）
		amqp.Publishing{
			Body: bytes,
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
	conn        *amqp.Connection
	Exchange    string
	DlxExchange string
}

func NewSubscriberManager(conn *amqp.Connection, exchange string, dlxExchange string) *SubscriberManagerImpl {
	ch, err := conn.Channel()
	if err != nil {
		klog.Fatal("NewChannel failed", err)
	}
	defer ch.Close()
	// 声明交换机
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
	}
	dlxch, err := conn.Channel()
	if err != nil {
		klog.Fatal("NewChannel failed", err)
	}
	defer dlxch.Close()
	// 声明死信交换机
	err = dlxch.ExchangeDeclare(
		config.GlobalServerConfig.RabbitMQ.DeadLetterExchange, // 死信交换机
		amqp.ExchangeTopic, // 交换机类型
		true,               // 是否持久化
		false,              // 是否自动删除
		false,              // 是否排他交换机
		false,              // 是否等待确认（即消息持久化）
		nil,                // 附加参数
	)
	if err != nil {
		klog.Fatal("ExchangeDeclare failed", err)
	}

	return &SubscriberManagerImpl{conn: conn, Exchange: exchange, DlxExchange: dlxExchange}
}

func (s *SubscriberManagerImpl) Consume(ctx context.Context, topic string, handler *ConsumerHandlerImpl) error {
	klog.Info("RabbitMQ 消费")
	msgs := s.SubscribeCh()
	if msgs == nil {
		klog.Fatal("SubscribeCh failed")
		return fmt.Errorf("SubscribeCh failed")
	}
	// 进行信息的消费。
	for {
		select {
		case msg := <-msgs:
			var Message base.InternalMessage
			err := sonic.Unmarshal(msg.Body, &Message)
			if err != nil {
				klog.Error("Unmarshal failed", err)
				continue
			}
		retry:
			tryNum := 0
			sf, err := snowflake.NewNode(consts.MessageIDSnowFlakeNode)
			if err != nil {
				klog.Error("NewNode failed", err)
				continue
			}
			id := sf.Generate().Int64()
			// 识别字段类型，确保正常处理。
			switch base.InternalMessageType(Message.Type) {
			case base.InternalMessageType_CHAT_MESSAGE:
				err := handler.MySQLManager.CreateChatMessage(context.Background(), &model.ChatMessage{
					ID:        id,
					SenderID:  Message.Payload.ChatMessage.UserId,
					RoomID:    Message.Payload.ChatMessage.RoomId,
					Content:   Message.Payload.ChatMessage.Message,
					CreatedAt: time.Now(),
				})
				if err != nil {
					klog.Error("CreateChatMessage failed", err)
					if tryNum < 3 {
						tryNum++
						goto retry
					}
					err = msg.Nack(false, true)
					if err != nil {
						klog.Error("Nack failed", err)
						continue
					}

				}
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

// 获取接受消息的通道
func (s *SubscriberManagerImpl) SubscribeCh() <-chan amqp.Delivery {
	ch, err := s.conn.Channel()
	if err != nil {
		klog.Fatal("NewChannel failed", err)
	}
	err = ch.ExchangeDeclare(
		s.Exchange,         // 交换机名称
		amqp.ExchangeTopic, // 交换机类型
		true,               // 是否持久化
		false,              // 是否自动删除
		false,              // 是否排他交换机
		false,              // 是否等待确认（即消息持久化）
		amqp.Table{
			"x-dead-letter-exchange":    config.GlobalServerConfig.RabbitMQ.DeadLetterExchange, // 死信交换机
			"x-dead-letter-routing-key": config.GlobalServerConfig.RabbitMQ.DeadLetterExchange, // 死信路由键
		},
	)
	if err != nil {
		klog.Fatal("ExchangeDeclare failed", err)
		return nil
	}
	queue, err := ch.QueueDeclare(
		"",    // 队列名称
		false, // 是否持久化
		false, // 是否排他
		false, // 是否自动删除
		false, // 是否阻塞
		nil,   // 附加参数
	)
	if err != nil {
		klog.Fatal("QueueDeclare failed", err)
		return nil
	}
	err = ch.QueueBind(
		queue.Name, // 队列名称
		"",
		s.Exchange, // 交换机名称
		false,      // 是否持久化
		nil,        // 附加参数
	)
	if err != nil {
		klog.Fatal("QueueBind failed", err)
		return nil
	}
	msgs, err := ch.Consume(
		queue.Name, // 队列名称
		"",         // 用来区分多个消费者的标识符
		false,      // 是否自动应答
		false,      // 是否消费者持久化
		false,      // 是否等待确认（即消息持久化）
		false,      // 是否排他
		nil,        // 附加参数
	)
	if err != nil {
		klog.Fatal("Consume failed", err)
		return nil
	}
	return msgs
}
