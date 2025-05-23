package mq

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/config"
	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/dao"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/base"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/quiz"
	"github.com/bytedance/sonic"
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
	for msg := range claim.Messages() {
		fmt.Append(msg.Value)
		panic("todo: handle message")
	}
	return nil
}
