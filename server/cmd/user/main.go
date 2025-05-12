package main

import (
	"log"

	"github.com/Rinai-R/ApexLecture/server/cmd/user/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/user/initialize"
	user "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/user/userservice"
)

func main() {
	initialize.InitConfig()
	db := initialize.InitDB()
	svr := user.NewServer(&UserServiceImpl{
		MysqlManager: dao.NewDM(db),
	})

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
