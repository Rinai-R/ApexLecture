package initialize

import (
	"net"

	"github.com/Rinai-R/ApexLecture/server/cmd/api/config"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/bwmarrin/snowflake"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	etcd "github.com/hertz-contrib/registry/etcd"
)

func InitRegistry() (registry.Registry, *registry.Info) {
	r, err := etcd.NewEtcdRegistry([]string{net.JoinHostPort(
		config.GlobalEtcdConfig.Host,
		config.GlobalEtcdConfig.Port,
	)})
	if err != nil {
		hlog.Fatal("failed to create etcd resolver: ", err)
	}
	suf, err := snowflake.NewNode(consts.EtcdSnowFlakeNode)
	if err != nil {
		hlog.Fatal("failed to create snowflake node: ", err)
	}
	info := &registry.Info{
		ServiceName: consts.ApiSrvPrefix,
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
	hlog.Info("initialize: registering service OK")
	return r, info
}
