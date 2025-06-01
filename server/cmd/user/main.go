package main

import (
	"context"
	"log"

	"github.com/Rinai-R/ApexLecture/server/cmd/user/config"
	"github.com/Rinai-R/ApexLecture/server/cmd/user/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/user/initialize"
	user "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/user/userservice"
	"github.com/cloudwego/kitex/pkg/limit"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/obs-opentelemetry/provider"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
)

func main() {
	initialize.Initlogger()
	initialize.InitConfig()

	db := initialize.InitDB()
	r, i := initialize.InitRegistry()
	private, public := initialize.InitKey()

	p := provider.NewOpenTelemetryProvider(
		provider.WithServiceName(config.GlobalServerConfig.Name),
		provider.WithExportEndpoint(config.GlobalServerConfig.OtelEndpoint),
		provider.WithInsecure(),
	)
	defer p.Shutdown(context.Background())
	svr := user.NewServer(&UserServiceImpl{
		MysqlManager: dao.NewDM(db),
		PrivateKey:   private,
		PublicKey:    public,
	},
		server.WithRegistry(r),
		server.WithRegistryInfo(i),
		server.WithServiceAddr(i.Addr),
		server.WithLimit(&limit.Option{MaxConnections: 2000, MaxQPS: 500}),
		server.WithSuite(tracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: config.GlobalServerConfig.Name,
		}),
	)

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
