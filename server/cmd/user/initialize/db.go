package initialize

import (
	"fmt"

	"github.com/Rinai-R/ApexLecture/server/cmd/user/config"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/cloudwego/kitex/pkg/klog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func InitDB() *gorm.DB {
	var db *gorm.DB
	var dsn string
	var err error

	dsn = fmt.Sprintf(consts.MysqlDNS,
		config.GlobalServerConfig.Mysql.Username,
		config.GlobalServerConfig.Mysql.Password,
		config.GlobalServerConfig.Mysql.Host,
		config.GlobalServerConfig.Mysql.Port,
		config.GlobalServerConfig.Mysql.Database,
	)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		klog.Fatal("initialize db failed: ", err)
	}
	klog.Info("initialize: initialize db success")
	return db
}
