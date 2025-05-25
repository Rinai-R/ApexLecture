package main

import (
	"context"
	"log"

	"github.com/Rinai-R/ApexLecture/server/cmd/agent/components/eino"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/config"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/initialize"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/agent/agentservice"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/hertz-contrib/obs-opentelemetry/provider"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
)

func main() {
	initialize.InitConfig()
	initialize.Initlogger()
	_, _ = initialize.InitMQ()
	r, i := initialize.InitRegistry()
	_ = initialize.InitDB()
	rdb := initialize.InitRedis()
	AskApp := initialize.InitEino()
	p := provider.NewOpenTelemetryProvider(
		provider.WithServiceName(config.GlobalServerConfig.Name),
		provider.WithExportEndpoint(config.GlobalServerConfig.OtelEndpoint),
		provider.WithInsecure(),
	)
	defer p.Shutdown(context.Background())

	svr := agentservice.NewServer(
		&AgentServiceImpl{
			RedisManager: dao.NewRedisManager(rdb),
			BotManager:   eino.NewBotManaer(AskApp),
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
