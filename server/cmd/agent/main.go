package main

import (
	"context"
	"log"

	"github.com/Rinai-R/ApexLecture/server/cmd/agent/components/eino"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/config"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/initialize"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/mq"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/agent/agentservice"
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
	initialize.InitConfig()
	initialize.Initlogger()
	conn := initialize.InitMqConn()
	r, i := initialize.InitRegistry()
	db := initialize.InitDB()
	m := initialize.InitMinio()
	rdb := initialize.InitRedis()
	AskApp := initialize.InitAskApp()
	SummaryApp := initialize.InitSummaryApp()
	handler := mq.NewConsumerHandler(dao.NewMysqlManager(db), m, eino.NewBotManaer(AskApp, SummaryApp))

	p := provider.NewOpenTelemetryProvider(
		provider.WithServiceName(config.GlobalServerConfig.Name),
		provider.WithExportEndpoint(config.GlobalServerConfig.OtelEndpoint),
		provider.WithInsecure(),
	)
	defer p.Shutdown(ctx)
	subscriber := mq.NewSubscriberManager(conn, config.GlobalServerConfig.RabbitMQ.Exchange, config.GlobalServerConfig.RabbitMQ.DeadLetterExchange)
	publisher := mq.NewPublisherManager(conn, config.GlobalServerConfig.RabbitMQ.Exchange)
	dlx_subscriber := mq.NewDLQConsumerManager(conn, config.GlobalServerConfig.RabbitMQ.DeadLetterExchange, "")
	go func() {
		err := subscriber.Consume(context.Background(), config.GlobalServerConfig.RabbitMQ.Exchange, handler)
		if err != nil {
			klog.Error("Consume failed", err)
		}
	}()
	go func() {
		err := dlx_subscriber.Consume(context.Background(), config.GlobalServerConfig.RabbitMQ.DeadLetterExchange, handler)
		if err != nil {
			klog.Error("DLX Consume failed", err)
		}
	}()
	svr := agentservice.NewServer(
		&AgentServiceImpl{
			RedisManager:    dao.NewRedisManager(rdb),
			BotManager:      eino.NewBotManaer(AskApp, SummaryApp),
			MysqlManager:    dao.NewMysqlManager(db),
			ProducerManager: publisher,
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
