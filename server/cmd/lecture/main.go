package main

import (
	"log"

	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/config"
	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/initialize"
	lecture "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/lecture/lectureservice"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
)

func main() {
	initialize.Initlogger()
	initialize.InitConfig()
	d := initialize.InitDB()
	r, i := initialize.InitRegistry()
	svr := lecture.NewServer(&LectureServiceImpl{
		MysqlManager: dao.NewDM(d),
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
