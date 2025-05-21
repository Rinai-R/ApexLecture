package main

import (
	"log"

	"github.com/Rinai-R/ApexLecture/server/cmd/push/config"
	"github.com/Rinai-R/ApexLecture/server/cmd/push/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/push/initialize"
	push "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/push/pushservice"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
)

func main() {
	initialize.InitConfig()
	initialize.Initlogger()
	r, i := initialize.InitRegistry()
	rdb := initialize.InitRedis()

	svr := push.NewServer(
		&PushServiceImpl{
			RedisManager: dao.NewRedisManager(rdb),
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
