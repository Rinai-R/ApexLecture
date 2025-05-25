package mq

import (
	"context"
	"fmt"
	"strconv"

	speech "cloud.google.com/go/speech/apiv1"
	"cloud.google.com/go/speech/apiv1/speechpb"
	"github.com/IBM/sarama"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/components/eino"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/config"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/model"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/minio/minio-go/v7"
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

			ctx := context.Background()
			object, err := h.MinioClient.GetObject(ctx,
				config.GlobalServerConfig.Minio.BucketName,
				fmt.Sprintf(consts.MinioObjectName, RoomId, "audio.ogg"),
				minio.GetObjectOptions{})
			if err != nil {
				klog.Error("MinioGetObject failed", err)
				continue
			}
			defer object.Close()
			var audioBytes []byte
			_, err = object.Read(audioBytes)
			if err != nil {
				klog.Error("ReadAll failed", err)
				continue
			}
			client, _ := speech.NewClient(ctx)
			resp, _ := client.Recognize(ctx, &speechpb.RecognizeRequest{
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
			text := resp.Results[0].Alternatives[0].Transcript

			SummarizedText := h.SummaryManager.Summary(ctx, &model.SummaryRequest{
				SummarizedText:   "",
				UnsummarizedText: text,
			})
			// 存储到 MySQL 数据库中
			err = h.MysqlManager.SetSummary(ctx, RoomId, SummarizedText.Summary)
			if err != nil {
				klog.Error("MysqlSetSummary failed", err)
				continue
			}
			session.MarkMessage(msg, "")
		case <-session.Context().Done():
			return nil
		}
	}
}
