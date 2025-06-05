package main

import (
	"context"
	"log"

	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/pkg/goroutine"
	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/config"
	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/initialize"
	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/mq"
	service "github.com/Rinai-R/ApexLecture/server/cmd/quiz/pkg/quiz_status"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/quiz/quizservice"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/limit"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/obs-opentelemetry/provider"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	initialize.Initlogger()
	initialize.InitConfig()
	d := initialize.InitDB()
	rdb := initialize.InitRedis()
	conn := initialize.InitMqConn()
	r, i := initialize.InitRegistry()
	handler := mq.NewConsumerHandler(dao.NewMysqlManager(d))
	p := provider.NewOpenTelemetryProvider(
		provider.WithServiceName(config.GlobalServerConfig.Name),
		provider.WithExportEndpoint(config.GlobalServerConfig.OtelEndpoint),
		provider.WithInsecure(),
	)
	defer p.Shutdown(ctx)
	subscriber := mq.NewSubscriberManager(conn, config.GlobalServerConfig.RabbitMQ.Exchange, config.GlobalServerConfig.RabbitMQ.DeadLetterExchange)
	publisher := mq.NewPublisherManager(conn, config.GlobalServerConfig.RabbitMQ.Exchange, config.GlobalServerConfig.RabbitMQ.DeadLetterExchange)
	DLQsubscriber := mq.NewDLQsubscriberManager(conn, config.GlobalServerConfig.RabbitMQ.DeadLetterExchange, "")
	go func() {
		err := subscriber.Consume(ctx, config.GlobalServerConfig.RabbitMQ.Exchange, handler)
		if err != nil {
			klog.Error("Consume failed", err)
		}
	}()
	go func() {
		err := DLQsubscriber.Consume(ctx, config.GlobalServerConfig.RabbitMQ.DeadLetterExchange, handler)
		if err != nil {
			klog.Error("DLQConsume failed", err)
		}
	}()
	svr := quizservice.NewServer(
		&QuizServiceImpl{
			MysqlManager:      dao.NewMysqlManager(d),
			RedisManager:      dao.NewRedisManager(rdb),
			ProducerManager:   publisher,
			goroutinePool:     goroutine.NewPool(1000),
			QuizStatusHanlder: service.NewQuizStatusHanlder(dao.NewRedisManager(rdb)),
		},
		server.WithRegistry(r),
		server.WithRegistryInfo(i),
		server.WithServiceAddr(i.Addr),
		server.WithLimit(&limit.Option{MaxConnections: 2000, MaxQPS: 500}),
		server.WithSuite(tracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: config.GlobalServerConfig.Name,
		}))
	err := svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}
