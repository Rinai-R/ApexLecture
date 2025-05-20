package rpc

import (
	"net"

	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/config"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/interaction/interactionservice"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/kitex-contrib/obs-opentelemetry/provider"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	etcd "github.com/kitex-contrib/registry-etcd"
)

func initInteraction() {
	r, err := etcd.NewEtcdResolver([]string{net.JoinHostPort(
		config.GlobalEtcdConfig.Host,
		config.GlobalEtcdConfig.Port,
	)})
	if err != nil {
		hlog.Fatal("initialize: failed to create etcd resolver", err)
	}
	provider.NewOpenTelemetryProvider(
		provider.WithServiceName(config.GlobalServerConfig.InteractionSrvInfo.Name),
		provider.WithExportEndpoint(config.GlobalServerConfig.OtelEndpoint),
		provider.WithInsecure(),
	)

	c, err := interactionservice.NewClient(
		config.GlobalServerConfig.InteractionSrvInfo.Name,
		client.WithResolver(r),
		client.WithSuite(tracing.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: config.GlobalServerConfig.InteractionSrvInfo.Name,
		}),
	)
	if err != nil {
		hlog.Fatal("initialize: failed to get interaction client ", err)
	}
	config.InteractionClient = c
	hlog.Info("initialize: interaction client initialized")
}
