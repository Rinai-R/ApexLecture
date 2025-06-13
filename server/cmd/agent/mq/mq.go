package mq

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	speech "cloud.google.com/go/speech/apiv1"
	"cloud.google.com/go/speech/apiv1/speechpb"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/components/eino"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/config"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/model"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/minio/minio-go/v7"
	"github.com/streadway/amqp"
	"google.golang.org/api/option"
)

type ConsumerManager interface {
	Consume(ctx context.Context, topic string, handler *ConsumerHandlerImpl) error
}

type ConsumerHandlerImpl struct {
	MysqlManager   *dao.MysqlManagerImpl
	MinioClient    *minio.Client
	SummaryManager *eino.BotManagerImpl
}

func NewConsumerHandler(MysqlManager *dao.MysqlManagerImpl, MinioClient *minio.Client, SummaryManager *eino.BotManagerImpl) *ConsumerHandlerImpl {
	return &ConsumerHandlerImpl{
		MysqlManager:   MysqlManager,
		MinioClient:    MinioClient,
		SummaryManager: SummaryManager,
	}
}

type PublisherManagerImpl struct {
	ch       *amqp.Channel
	Exchange string
}

func NewPublisherManager(conn *amqp.Connection, exchange string) *PublisherManagerImpl {
	ch, err := conn.Channel()
	if err != nil {
		klog.Fatal("NewChannel failed", err)
		return nil
	}

	err = ch.ExchangeDeclare(
		exchange,           // 交换机名称
		amqp.ExchangeTopic, // 交换机类型
		true,               // 是否持久化
		false,              // 是否自动删除
		false,              // 是否内部使用
		false,              // 是否排外
		nil,                // 额外参数
	)
	if err != nil {
		klog.Fatal("ExchangeDeclare failed", err)
		return nil
	}

	klog.Info("RabbitMQ 消费准备就绪")
	return &PublisherManagerImpl{ch: ch, Exchange: exchange}
}

func (p *PublisherManagerImpl) Send(ctx context.Context, RoomId int64) error {
	err := p.ch.Publish(
		p.Exchange, // 交换机名称
		"",
		false, // 是否强制路由
		false, // 是否立即发送
		amqp.Publishing{
			Body: []byte(fmt.Sprintf("%d", RoomId)),
		},
	)
	if err != nil {
		klog.Error("Publish failed", err)
		return err
	}
	klog.Info("RabbitMQ 发布成功")
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

	return &SubscriberManagerImpl{
		conn:               conn,
		Exchange:           exchange,
		DeadLetterExchange: dlxExchange,
	}
}

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
			"x-dead-letter-exchange":    s.DeadLetterExchange, // 死信交换机
			"x-dead-letter-routing-key": "",                   // 死信路由键
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

func (s *SubscriberManagerImpl) Consume(ctx context.Context, _ string, h *ConsumerHandlerImpl) error {
	msgs := s.SubscribeCh()
	if msgs == nil {
		klog.Fatal("SubscribeCh failed")
		return fmt.Errorf("SubscribeCh failed")
	}
	for msg := range msgs {
		tryNum := 0
	retry:
		RoomId, err := strconv.ParseInt(string(msg.Body), 10, 64)
		if err != nil {
			// 说明传递的消息格式错误，丢弃
			klog.Warn("RabbitMQConsumer:ParseInt failed", err)
			msg.Ack(false)
			continue
		}
		klog.Info("RabbitMQ 消费", RoomId)
		object, err := h.MinioClient.GetObject(ctx,
			config.GlobalServerConfig.Minio.BucketName,
			fmt.Sprintf(consts.MinioObjectName, RoomId, "audio.ogg"),
			minio.GetObjectOptions{})
		if err != nil {
			// 说明获取对象失败，可能是对象不存在或者其他错误，丢弃
			klog.Warn("MinioGetObject failed", err)
			msg.Ack(false)
			continue
		}

		var audioBytes []byte
		_, err = object.Read(audioBytes)
		if err != nil {
			klog.Warn("ReadAll failed", err)
			msg.Nack(false, false)
			continue
		}
		// 此处凭证需要自行添加。
		// 有凭证之后理论可以实现智能纪要
		client, err := speech.NewClient(ctx,
			option.WithCredentialsFile(consts.GoogleCredentialsFile),
		)
		if err != nil {
			klog.Error("SpeechNewClient failed", err)
			if tryNum < 3 {
				tryNum++
				goto retry
			}
			klog.Warn("tryNum > 3, SpeechNewClient failed", err)
			msg.Nack(false, false)
			continue
		}
		resp, err := client.Recognize(ctx, &speechpb.RecognizeRequest{
			Audio: &speechpb.RecognitionAudio{
				AudioSource: &speechpb.RecognitionAudio_Content{
					Content: audioBytes,
				},
			},
			Config: &speechpb.RecognitionConfig{
				Encoding:                   speechpb.RecognitionConfig_OGG_OPUS,
				LanguageCode:               "zh-CN",
				EnableAutomaticPunctuation: true,
			},
		})
		if err != nil {
			klog.Error("Recognize failed", err)
			if tryNum < 3 {
				tryNum++
				goto retry
			}
			klog.Warn("tryNum > 3, Recognize failed", err)
			msg.Nack(false, false)
			continue
		}
		text := resp.Results[0].Alternatives[0].Transcript

		SummarizedText := h.SummaryManager.Summary(ctx, &model.SummaryRequest{
			SummarizedText:   "",
			UnsummarizedText: text,
		})
		// 存储到 MySQL 数据库中
		err = h.MysqlManager.SetSummary(ctx, RoomId, SummarizedText.Summary)
		if err != nil {
			if strings.Contains(err.Error(), "Error 1062") {
				klog.Warn("Duplicate entry", err)
				msg.Ack(false)
				continue
			}
			klog.Error("MysqlSetSummary failed", err)
			if tryNum < 3 {
				tryNum++
				goto retry
			}
			klog.Error("tryNum > 3, MysqlSetSummary failed", err)
			msg.Nack(false, false)
			continue
		}
		msg.Ack(false)
		object.Close()
	}
	return nil
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
		"",    // 队列名称
		false, // 是否持久化
		false, // 是否排他
		false, // 是否自动删除
		false, // 是否阻塞
		nil,   // 附加参数
	)
	if err != nil {
		klog.Fatal("QueueDeclare failed", err)
		return fmt.Errorf("QueueDeclare failed")
	}

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
