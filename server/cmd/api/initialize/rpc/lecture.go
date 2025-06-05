package rpc

import (
	"net"

	"github.com/Rinai-R/ApexLecture/server/cmd/api/config"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/lecture/lectureservice"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/kitex-contrib/obs-opentelemetry/provider"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	etcd "github.com/kitex-contrib/registry-etcd"
)

func initLecture() {
	r, err := etcd.NewEtcdResolver([]string{net.JoinHostPort(
		config.GlobalEtcdConfig.Host,
		config.GlobalEtcdConfig.Port,
	)})
	if err != nil {
		hlog.Fatal("initialize: failed to create etcd resolver", err)
	}
	provider.NewOpenTelemetryProvider(
		provider.WithServiceName(config.GlobalServerConfig.LectureSrvInfo.Name),
		provider.WithExportEndpoint(config.GlobalServerConfig.OtelEndpoint),
		provider.WithInsecure(),
	)
	// 一致性哈希，确保同一个房间的请求都落在同一个节点上
	// f := func(ctx context.Context, request interface{}) string {
	// 	return ctx.Value("roomid").(string)
	// }
	// balancer := loadbalance.ConsistentHashOption{
	// 	GetKey:         f,
	// 	Replica:        0,
	// 	VirtualFactor:  100,
	// 	Weighted:       true,
	// 	ExpireDuration: time.Minute * 2,
	// }
	c, err := lectureservice.NewClient(
		config.GlobalServerConfig.LectureSrvInfo.Name,
		client.WithResolver(r),
		client.WithSuite(tracing.NewClientSuite()),
		// client.WithLoadBalancer(loadbalance.NewConsistBalancer(balancer)),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: config.GlobalServerConfig.LectureSrvInfo.Name,
		}),
	)
	if err != nil {
		hlog.Fatal("initialize: failed to get lecture client ", err)
	}
	config.LectureClient = c
	hlog.Info("initialize: lecture client initialized")
}
