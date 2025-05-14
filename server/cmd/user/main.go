package main

import (
	"log"

	"github.com/Rinai-R/ApexLecture/server/cmd/user/config"
	"github.com/Rinai-R/ApexLecture/server/cmd/user/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/user/initialize"
	user "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/user/userservice"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
)

func main() {
	initialize.Initlogger()
	initialize.InitConfig()

	db := initialize.InitDB()
	r, i := initialize.InitRegistry()
	private, public := initialize.InitKey()

	svr := user.NewServer(&UserServiceImpl{
		MysqlManager: dao.NewDM(db),
		PrivateKey:   private,
		PublicKey:    public,
	},
		server.WithRegistry(r),
		server.WithRegistryInfo(i),
		server.WithServiceAddr(i.Addr),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: config.GlobalServerConfig.Name,
		}),
	)

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
