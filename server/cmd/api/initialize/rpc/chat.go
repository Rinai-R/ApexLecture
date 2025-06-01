package rpc

import (
	"net"

	"github.com/Rinai-R/ApexLecture/server/cmd/api/config"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/chat/chatservice"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/loadbalance"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/kitex-contrib/obs-opentelemetry/provider"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	etcd "github.com/kitex-contrib/registry-etcd"
)

func initChat() {
	r, err := etcd.NewEtcdResolver([]string{net.JoinHostPort(
		config.GlobalEtcdConfig.Host,
		config.GlobalEtcdConfig.Port,
	)})
	if err != nil {
		hlog.Fatal("initialize: failed to create etcd resolver", err)
	}
	provider.NewOpenTelemetryProvider(
		provider.WithServiceName(config.GlobalServerConfig.ChatSrvInfo.Name),
		provider.WithExportEndpoint(config.GlobalServerConfig.OtelEndpoint),
		provider.WithInsecure(),
	)

	c, err := chatservice.NewClient(
		config.GlobalServerConfig.ChatSrvInfo.Name,
		client.WithResolver(r),
		client.WithSuite(tracing.NewClientSuite()),
		client.WithLoadBalancer(loadbalance.NewInterleavedWeightedRoundRobinBalancer()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: config.GlobalServerConfig.ChatSrvInfo.Name,
		}),
	)
	if err != nil {
		hlog.Fatal("initialize: failed to get chat client ", err)
	}
	config.ChatClient = c
	hlog.Info("initialize: chat client initialized")
}
