package mq

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/config"
	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/model"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/base"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/quiz"
	"github.com/bwmarrin/snowflake"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/kitex/pkg/klog"
	"gorm.io/gorm"
)

type ProducerManagerImpl struct {
	producer sarama.AsyncProducer
}

func NewProducerManager(producer sarama.AsyncProducer) *ProducerManagerImpl {
	return &ProducerManagerImpl{
		producer: producer,
	}
}

func (p *ProducerManagerImpl) ProduceQuestion(ctx context.Context, req *quiz.SubmitQuestionRequest, questionId int64) error {
	var msg *base.InternalMessage
	// 根据类型进行序列化
	switch base.InternalMessageType(req.Type) {
	case base.InternalMessageType_QUIZ_CHOICE:
		msg = &base.InternalMessage{
			Type: base.InternalMessageType(req.Type),
			Payload: &base.InternalPayload{
				QuizChoice: &base.InternalQuizChoice{
					QuestionId: questionId,
					UserId:     req.UserId,
					RoomId:     req.RoomId,
					Title:      req.Payload.Choice.Title,
					Options:    req.Payload.Choice.Options,
					Answers:    req.Payload.Choice.Answers,
				},
			},
		}
	case base.InternalMessageType_QUIZ_JUDGE:
		msg = &base.InternalMessage{
			Type: base.InternalMessageType(req.Type),
			Payload: &base.InternalPayload{
				QuizJudge: &base.InternalQuizJudge{
					QuestionId: questionId,
					UserId:     req.UserId,
					RoomId:     req.RoomId,
					Title:      req.Payload.Judge.Title,
					Answer:     req.Payload.Judge.Answer,
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
		Key:   sarama.StringEncoder(fmt.Sprintf(consts.RoomKey, req.RoomId)),
		Value: sarama.StringEncoder(bytes),
	}
	return nil
}

type ConsumerManagerImpl struct {
	consumer sarama.ConsumerGroup
}

func NewConsumerManager(consumer sarama.ConsumerGroup) *ConsumerManagerImpl {
	return &ConsumerManagerImpl{
		consumer: consumer,
	}
}

func (c *ConsumerManagerImpl) Consume(ctx context.Context, topics string, handler *ConsumerHandlerImpl) error {
	return c.consumer.Consume(ctx, []string{topics}, handler)
}

type ConsumerHandlerImpl struct {
	MysqlManager *dao.MysqlManagerImpl
}

func NewConsumerHandler(MysqlManager *dao.MysqlManagerImpl) *ConsumerHandlerImpl {
	return &ConsumerHandlerImpl{
		MysqlManager: MysqlManager,
	}
}

func (h *ConsumerHandlerImpl) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerHandlerImpl) Cleanup(session sarama.ConsumerGroupSession) error {
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
			case base.InternalMessageType_QUIZ_CHOICE:
				err := h.MysqlManager.CreateQuizChoice(&model.QuizChoice{
					Id:     id,
					RoomId: Message.Payload.QuizChoice.RoomId,
					UserId: Message.Payload.QuizChoice.UserId,
					Title:  Message.Payload.QuizChoice.Title,
					Option: Message.Payload.QuizChoice.Options,
					Answer: Message.Payload.QuizChoice.Answers,
				})
				if err != nil {
					if err == gorm.ErrDuplicatedKey {
						klog.Error("QuizChoice already exists", err)
					} else {
						klog.Error("CreateQuizChoice failed", err)
					}
					continue
				}
				session.MarkMessage(message, "")

			case base.InternalMessageType_QUIZ_JUDGE:
				err := h.MysqlManager.CreateQuizJudge(&model.QuizJudge{
					Id:     id,
					RoomId: Message.Payload.QuizJudge.RoomId,
					UserId: Message.Payload.QuizJudge.UserId,
					Title:  Message.Payload.QuizJudge.Title,
					Answer: Message.Payload.QuizJudge.Answer,
				})
				if err != nil {
					if err == gorm.ErrDuplicatedKey {
						klog.Error("QuizChoice already exists", err)
					} else {
						klog.Error("CreateQuizChoice failed", err)
					}
					continue
				}
				session.MarkMessage(message, "")
			default:
				klog.Error("Unknown message type", base.InternalMessageType(Message.Type))
			}
		case <-session.Context().Done():
			return nil
		}
	}
}
