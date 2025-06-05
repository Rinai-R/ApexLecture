package main

import (
	"context"
	"log"

	"github.com/Rinai-R/ApexLecture/server/cmd/push/config"
	"github.com/Rinai-R/ApexLecture/server/cmd/push/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/push/initialize"
	"github.com/Rinai-R/ApexLecture/server/cmd/push/pkg/sensitive"
	push "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/push/pushservice"
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
	r, i := initialize.InitRegistry()
	rdb := initialize.InitRedis()
	p := provider.NewOpenTelemetryProvider(
		provider.WithServiceName(config.GlobalServerConfig.Name),
		provider.WithExportEndpoint(config.GlobalServerConfig.OtelEndpoint),
		provider.WithInsecure(),
	)
	defer p.Shutdown(ctx)

	svr := push.NewServer(
		&PushServiceImpl{
			RedisManager: dao.NewRedisManager(rdb),
			Filter:       sensitive.NewFilter(),
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
