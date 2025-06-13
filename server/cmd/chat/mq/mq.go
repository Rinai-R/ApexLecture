package mq

import (
	"context"
	"fmt"
	"strings"
	"time"

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

type ConsumerManager interface {
	Consume(ctx context.Context, topic string, handler *ConsumerHandlerImpl) error
}

type ConsumerHandlerImpl struct {
	LoveFrom5mm  string
	MySQLManager *dao.MysqlManagerImpl
}

func NewConsumerHandler(mysqlManager *dao.MysqlManagerImpl) *ConsumerHandlerImpl {
	return &ConsumerHandlerImpl{
		LoveFrom5mm:  "曹寺5mm",
		MySQLManager: mysqlManager,
	}
}

// ============================= 兔子 mq ========================

type PublisherManagerImpl struct {
	ch                 *amqp.Channel
	Exchange           string
	DeadLetterExchange string
}

func NewPublisherManager(conn *amqp.Connection, exchange, deadLetterExchange string) *PublisherManagerImpl {
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
		amqp.Table{
			"x-dead-letter-exchange":    deadLetterExchange, // 死信交换机
			"x-dead-letter-routing-key": deadLetterExchange, // 死信路由键
		},
	)
	if err != nil {
		klog.Fatal("ExchangeDeclare failed", err)
		return nil
	}
	return &PublisherManagerImpl{ch: ch, Exchange: exchange, DeadLetterExchange: deadLetterExchange}
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
		nil,
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
		dlxExchange,        // 死信交换机
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

func (s *SubscriberManagerImpl) Consume(ctx context.Context, _ string, handler *ConsumerHandlerImpl) error {
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
					if strings.Contains(err.Error(), "Error 1062") {
						klog.Warn("Duplicate entry", err)
						msg.Ack(false)
						continue
					}
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
		nil,
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
		amqp.Table{
			"x-dead-letter-exchange":    s.DlxExchange, // 死信交换机
			"x-dead-letter-routing-key": "",            // 死信路由键，这里统一管理就滞空了。
		},
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

type DLQConsumerManager struct {
	conn               *amqp.Connection
	DeadLetterExchange string
	Exchange           string
}

func NewDLQConsumerManager(conn *amqp.Connection, deadLetterExchange, Exchange string) *DLQConsumerManager {
	ch, err := conn.Channel()
	if err != nil {
		klog.Fatal("NewChannel failed", err)
		return nil
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		deadLetterExchange, // 死信交换机名称
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

	return &DLQConsumerManager{
		conn:               conn,
		DeadLetterExchange: deadLetterExchange,
		Exchange:           Exchange,
	}
}

func (d *DLQConsumerManager) Consume(ctx context.Context, _ string, h *ConsumerHandlerImpl) error {
	ch, err := d.conn.Channel()
	if err != nil {
		klog.Fatal("NewChannel failed", err)
		return fmt.Errorf("NewChannel failed")
	}
	defer ch.Close()

	queue, err := ch.QueueDeclare(
		d.DeadLetterExchange, // 队列名称
		false,                // 是否持久化
		false,                // 是否排他
		false,                // 是否自动删除
		false,                // 是否阻塞
		nil,                  // 附加参数
	)
	if err != nil {
		klog.Fatal("QueueDeclare failed", err)
		return fmt.Errorf("QueueDeclare failed")
	}

	err = ch.QueueBind(
		queue.Name,           // 队列名称
		"",                   // 路由键
		d.DeadLetterExchange, // 死信交换机名称
		false,                // 是否持久化
		nil,                  // 附加参数
	)
	if err != nil {
		klog.Fatal("QueueBind failed", err)
		return fmt.Errorf("QueueBind failed")
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
		return fmt.Errorf("Consume failed")
	}
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		select {
		case msg := <-msgs:
			klog.Error("DLQ Consume: ", string(msg.Body))
			// 这里重新发给队列
			ch.Publish(
				d.Exchange, // 交换机名称
				"",         // 路由键
				false,      // 是否强制路由
				false,      // 是否立即发送
				amqp.Publishing{
					Body: msg.Body,
				},
			)
		case <-ticker.C:
			klog.Debug("DLQ Consumer heartbeat, Current DLQ number: ", len(msgs))
		case <-ctx.Done():
			klog.Info("DLQ Consumer stopped")
			return nil
		}
	}
}
