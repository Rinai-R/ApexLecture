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
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/hertz-contrib/obs-opentelemetry/provider"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
)

func main() {
	initialize.InitConfig()
	initialize.Initlogger()
	pro, con := initialize.InitMQ()
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
	defer p.Shutdown(context.Background())
	go func() {
		consumer := mq.NewConsumerManager(con)
		err := consumer.Consume(context.Background(), config.GlobalServerConfig.Kafka.Topic, handler)
		if err != nil {
			klog.Error("Consume failed", err)
		}
	}()
	svr := agentservice.NewServer(
		&AgentServiceImpl{
			RedisManager:    dao.NewRedisManager(rdb),
			BotManager:      eino.NewBotManaer(AskApp, SummaryApp),
			MysqlManager:    dao.NewMysqlManager(db),
			ProducerManager: mq.NewProducerManager(pro),
		},
		server.WithRegistry(r),
		server.WithRegistryInfo(i),
		server.WithServiceAddr(i.Addr),
		server.WithSuite(tracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: config.GlobalServerConfig.Name,
		}))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
