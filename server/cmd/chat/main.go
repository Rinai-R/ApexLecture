package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Rinai-R/ApexLecture/server/cmd/chat/config"
	"github.com/Rinai-R/ApexLecture/server/cmd/chat/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/chat/initialize"
	"github.com/Rinai-R/ApexLecture/server/cmd/chat/mq"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/chat/chatservice"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/obs-opentelemetry/provider"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
)

func main() {
	initialize.Initlogger()
	initialize.InitConfig()
	d := initialize.InitDB()
	rdb := initialize.InitRedis()
	pro, con := initialize.InitMQ()
	r, i := initialize.InitRegistry()
	handler := mq.NewConsumerHandler(dao.NewMysqlManager(d))
	p := provider.NewOpenTelemetryProvider(
		provider.WithServiceName(config.GlobalServerConfig.Name),
		provider.WithExportEndpoint(config.GlobalServerConfig.OtelEndpoint),
		provider.WithInsecure(),
	)
	defer p.Shutdown(context.Background())
	svr := chatservice.NewServer(
		&ChatServiceImpl{
			MysqlManager: dao.NewMysqlManager(d),
			RedisManager: dao.NewRedisManager(rdb),
			MQManager:    mq.NewProducerManager(pro),
		},
		server.WithRegistry(r),
		server.WithRegistryInfo(i),
		server.WithServiceAddr(i.Addr),
		server.WithSuite(tracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: config.GlobalServerConfig.Name,
		}))
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		consumer := mq.NewConsumerManager(con)
		err := consumer.Consume(ctx, config.GlobalServerConfig.Kafka.Topic, handler)
		if err != nil {
			klog.Error("Consume failed", err)
		}
	}()
	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		<-signalChan
		cancel()
	}()
	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
