package mq

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/model"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/base"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/quiz"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/streadway/amqp"
)

type ConsumerHandlerImpl struct {
	MysqlManager *dao.MysqlManagerImpl
}

func NewConsumerHandler(MysqlManager *dao.MysqlManagerImpl) *ConsumerHandlerImpl {
	return &ConsumerHandlerImpl{
		MysqlManager: MysqlManager,
	}
}

type PublisherManagerImpl struct {
	ch                 *amqp.Channel
	Exchange           string
	DeadLetterExchange string
}

func NewPublisherManager(conn *amqp.Connection, exchange, DeadLetterExchange string) *PublisherManagerImpl {
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
	return &PublisherManagerImpl{ch: ch, Exchange: exchange, DeadLetterExchange: DeadLetterExchange}
}

func (p *PublisherManagerImpl) ProduceQuestion(ctx context.Context, request *quiz.SubmitQuestionRequest, questionId int64) error {
	var msg *base.InternalMessage
	// 根据类型进行序列化
	switch base.InternalMessageType(request.Type) {
	case base.InternalMessageType_QUIZ_CHOICE:
		msg = &base.InternalMessage{
			Type: base.InternalMessageType(request.Type),
			Payload: &base.InternalPayload{
				QuizChoice: &base.InternalQuizChoice{
					QuestionId: questionId,
					UserId:     request.UserId,
					RoomId:     request.RoomId,
					Title:      request.Payload.Choice.Title,
					Options:    request.Payload.Choice.Options,
					Answers:    request.Payload.Choice.Answers,
				},
			},
		}
	case base.InternalMessageType_QUIZ_JUDGE:
		msg = &base.InternalMessage{
			Type: base.InternalMessageType(request.Type),
			Payload: &base.InternalPayload{
				QuizJudge: &base.InternalQuizJudge{
					QuestionId: questionId,
					UserId:     request.UserId,
					RoomId:     request.RoomId,
					Title:      request.Payload.Judge.Title,
					Answer:     request.Payload.Judge.Answer,
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
		"",
		false, // 是否强制路由
		false, // 是否等待确认
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

type SubscriberManagerImpl struct {
	conn               *amqp.Connection
	Exchange           string
	DeadLetterExchange string
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

	return &SubscriberManagerImpl{conn: conn, Exchange: exchange, DeadLetterExchange: dlxExchange}
}

func (s *SubscriberManagerImpl) Consume(ctx context.Context, queue string, handler *ConsumerHandlerImpl) error {
	ch, err := s.conn.Channel()
	if err != nil {
		klog.Error("NewChannel failed", err)
		return err
	}
	defer ch.Close()
	// 声明队列
	q, err := ch.QueueDeclare(
		queue, // 队列名称
		true,  // 是否持久化
		false, // 是否排他
		false, // 是否自动删除
		false, // 是否等待确认（即消息持久化）
		amqp.Table{
			"x-dead-letter-exchange":    s.DeadLetterExchange, // 死信交换机
			"x-dead-letter-routing-key": "",                   // 死信路由键
		},
	)
	if err != nil {
		klog.Error("QueueDeclare failed", err)
		return err
	}
	// 绑定队列到交换机
	err = ch.QueueBind(
		q.Name,     // 队列名称
		"",         // 路由键
		s.Exchange, // 交换机名称
		false,      // 是否等待确认
		nil,        // 附加参数
	)
	if err != nil {
		klog.Error("QueueBind failed", err)
		return err
	}
	// 声明消费者
	consumer, err := ch.Consume(
		q.Name, // 队列名称
		"",     // 用来区分多个消费者的标识符
		false,  // 是否自动确认
		false,  // 是否 exclusive
		false,  // 是否 no-local
		false,  // 是否 no-wait
		nil,    // 其他参数
	)
	if err != nil {
		klog.Error("Consume failed", err)
		return err
	}
	for {
		select {
		case message := <-consumer:
			retryCount := 0
		retry:
			fmt.Println(string(message.Body))
			var Message base.InternalMessage
			err := sonic.Unmarshal(message.Body, &Message)
			if err != nil {
				klog.Error("Unmarshal failed", err)
				continue
			}
			switch base.InternalMessageType(Message.Type) {
			case base.InternalMessageType_QUIZ_CHOICE:
				err := handler.MysqlManager.CreateQuizChoice(&model.QuizChoice{
					Id:     Message.Payload.QuizChoice.QuestionId,
					RoomId: Message.Payload.QuizChoice.RoomId,
					UserId: Message.Payload.QuizChoice.UserId,
					Title:  Message.Payload.QuizChoice.Title,
					Option: Message.Payload.QuizChoice.Options,
					Answer: Message.Payload.QuizChoice.Answers,
				})
				if err != nil {
					if strings.Contains(err.Error(), "Error 1062") {
						klog.Warn("Duplicate entry", err)
						message.Ack(false)
						continue
					}
					klog.Error("CreateQuizChoice failed, ", err)
					if retryCount < 3 {
						retryCount++
						goto retry
					}
					message.Nack(false, false)
					continue
				}
			case base.InternalMessageType_QUIZ_JUDGE:
				err := handler.MysqlManager.CreateQuizJudge(&model.QuizJudge{
					Id:     Message.Payload.QuizJudge.QuestionId,
					RoomId: Message.Payload.QuizJudge.RoomId,
					UserId: Message.Payload.QuizJudge.UserId,
					Title:  Message.Payload.QuizJudge.Title,
					Answer: Message.Payload.QuizJudge.Answer,
				})
				if err != nil {
					if strings.Contains(err.Error(), "Error 1062") {
						klog.Warn("Duplicate entry", err)
						message.Ack(false)
						continue
					}
					klog.Error("CreateQuizJudge failed, ", err)
					if retryCount < 3 {
						retryCount++
						goto retry
					}
					message.Nack(false, false)
				}
			default:
				klog.Error("Unknown message type", base.InternalMessageType(Message.Type))
			}
			message.Ack(false)
			if err != nil {
				klog.Error("Ack failed", err)
				continue
			}
		case <-ctx.Done():
			return nil
		}
	}
}

type DLQConsumerManagerImpl struct {
	conn               *amqp.Connection
	Exchange           string
	DeadLetterExchange string
}

func NewDLQsubscriberManager(conn *amqp.Connection, exchange string, dlxExchange string) *DLQConsumerManagerImpl {
	ch, err := conn.Channel()
	if err != nil {
		klog.Fatal("NewChannel failed", err)
	}
	defer ch.Close()
	// 声明死信交换机
	err = ch.ExchangeDeclare(
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

	return &DLQConsumerManagerImpl{conn: conn, Exchange: exchange, DeadLetterExchange: dlxExchange}
}

func (d *DLQConsumerManagerImpl) Consume(ctx context.Context, _ string, h *ConsumerHandlerImpl) error {
	ch, err := d.conn.Channel()
	if err != nil {
		klog.Fatal("NewChannel failed", err)
		return fmt.Errorf("NewChannel failed")
	}
	defer ch.Close()
	// 声明队列，使用空字符串，每次都会得到不同的 name
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
	// 将我们获取的队列绑定到死信交换机
	err = ch.QueueBind(
		queue.Name, // 队列名称
		"",
		d.DeadLetterExchange, // 死信交换机名称
		false,                // 是否持久化
		nil,                  // 附加参数
	)
	if err != nil {
		klog.Fatal("QueueBind failed", err)
		return fmt.Errorf("QueueBind failed")
	}
	// 这里是进行消费队列，获取消息
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
			klog.Info("DLQ Consumer heartbeat, Current DLQ number: ", len(msgs))
		case <-ctx.Done():
			klog.Info("DLQ Consumer stopped")
			return nil
		}
	}
}
