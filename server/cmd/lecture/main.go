package main

import (
	"context"
	"log"
	"sync"

	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/config"
	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/initialize"
	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/pkg/goroutine"
	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/pkg/webrtc"
	lecture "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/lecture/lectureservice"
	"github.com/cloudwego/kitex/pkg/limit"
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
	r, i := initialize.InitRegistry()
	m := initialize.InitMinio()
	p := provider.NewOpenTelemetryProvider(
		provider.WithServiceName(config.GlobalServerConfig.Name),
		provider.WithExportEndpoint(config.GlobalServerConfig.OtelEndpoint),
		provider.WithInsecure(),
	)
	defer p.Shutdown(context.Background())
	svr := lecture.NewServer(&LectureServiceImpl{
		MysqlManager:         dao.NewDM(d),
		RedisManager:         dao.NewRedisManager(rdb),
		MinioManager:         m,
		Sessions:             &sync.Map{},
		WebrtcAPI:            webrtc.NewWebrtcAPI(),
		peerConnectionConfig: webrtc.WebrtcConfig(),
		goroutinePool:        goroutine.NewPool(1000),
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
