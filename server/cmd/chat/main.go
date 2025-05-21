package main

import (
	"context"
	"log"

	"github.com/Rinai-R/ApexLecture/server/cmd/chat/config"
	"github.com/Rinai-R/ApexLecture/server/cmd/chat/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/chat/initialize"
	"github.com/Rinai-R/ApexLecture/server/cmd/chat/mq"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/chat/chatservice"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/hertz-contrib/obs-opentelemetry/provider"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
)

func main() {
	initialize.Initlogger()
	initialize.InitConfig()
	d := initialize.InitDB()
	rdb := initialize.InitRedis()
	pro, con := initialize.InitMQ()
	r, i := initialize.InitRegistry()
	p := provider.NewOpenTelemetryProvider(
		provider.WithServiceName(config.GlobalServerConfig.Name),
		provider.WithExportEndpoint(config.GlobalServerConfig.OtelEndpoint),
		provider.WithInsecure(),
	)
	defer p.Shutdown(context.Background())
	svr := chatservice.NewServer(
		&ChatServiceImpl{
			MysqlManagerImpl: dao.NewMysqlManager(d),
			RedisManagerImpl: dao.NewRedisManager(rdb),
			MQManagerImpl:    mq.NewProducerManager(pro),
		},
		server.WithRegistry(r),
		server.WithRegistryInfo(i),
		server.WithServiceAddr(i.Addr),
		server.WithSuite(tracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: config.GlobalServerConfig.Name,
		}))
	go func() {
		consumer := mq.NewConsumerManager(con)
		err := consumer.Consume(context.Background(), config.GlobalServerConfig.Kafka.Topic)
		if err != nil {
			klog.Error("Consume failed", err)
		}
	}()
	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
