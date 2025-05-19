package initialize

import (
	"net"

	"github.com/Rinai-R/ApexLecture/server/cmd/interaction/config"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/bwmarrin/snowflake"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/pkg/utils"
	etcd "github.com/kitex-contrib/registry-etcd"
)

func InitRegistry() (registry.Registry, *registry.Info) {
	r, err := etcd.NewEtcdRegistry([]string{net.JoinHostPort(
		config.GlobalEtcdConfig.Host,
		config.GlobalEtcdConfig.Port,
	)})
	if err != nil {
		klog.Fatal("failed to create etcd resolver: ", err)
	}
	suf, err := snowflake.NewNode(consts.EtcdSnowFlakeNode)
	if err != nil {
		klog.Fatal("failed to create snowflake node: ", err)
	}
	info := &registry.Info{
		ServiceName: config.GlobalServerConfig.Name,
		Addr: utils.NewNetAddr(
			"tcp",
			net.JoinHostPort(
				config.GlobalServerConfig.Host,
				config.GlobalServerConfig.Port,
			),
		),
		Tags: map[string]string{
			"ID": suf.Generate().Base36(),
		},
	}
	klog.Info("initialize: registering service OK")
	return r, info
}
