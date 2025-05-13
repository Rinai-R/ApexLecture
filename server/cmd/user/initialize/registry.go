package initialize

import (
	"net"

	"github.com/Rinai-R/ApexLecture/server/cmd/user/config"
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
		klog.Fatalf("failed to create etcd resolver: %v", err)
	}
	suf, err := snowflake.NewNode(consts.UserSrvSnowFlakeNode)
	if err != nil {
		klog.Fatalf("failed to create snowflake node: %v", err)
	}
	info := &registry.Info{
		ServiceName: consts.UserSrvPrefix,
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
	klog.Infof("initialize: registering service OK")
	return r, info
}
