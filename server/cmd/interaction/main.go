package main

import (
	"context"
	"log"

	"github.com/Rinai-R/ApexLecture/server/cmd/interaction/config"
	"github.com/Rinai-R/ApexLecture/server/cmd/interaction/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/interaction/initialize"
	interaction "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/interaction/interactionservice"
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
	r, i := initialize.InitRegistry()
	p := provider.NewOpenTelemetryProvider(
		provider.WithServiceName(config.GlobalServerConfig.Name),
		provider.WithExportEndpoint(config.GlobalServerConfig.OtelEndpoint),
		provider.WithInsecure(),
	)
	defer p.Shutdown(context.Background())
	svr := interaction.NewServer(
		&InteractionServiceImpl{
			MysqlManagerImpl: dao.NewMysqlManager(d),
			RedisManagerImpl: dao.NewRedisManager(rdb),
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
