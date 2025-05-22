package initialize

import (
	"fmt"

	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/config"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func InitDB() *gorm.DB {
	dsn := fmt.Sprintf(consts.MysqlDNS,
		config.GlobalServerConfig.Mysql.Username,
		config.GlobalServerConfig.Mysql.Password,
		config.GlobalServerConfig.Mysql.Host,
		config.GlobalServerConfig.Mysql.Port,
		config.GlobalServerConfig.Mysql.Database,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
		Logger:         logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		klog.Fatal("Failed to connect to mysql", err)
	}
	klog.Info("Connected to mysql")
	return db
}

func InitRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf(
			"%s:%s",
			config.GlobalServerConfig.Redis.Host,
			config.GlobalServerConfig.Redis.Port,
		),
		Password: config.GlobalServerConfig.Redis.Password,
		DB:       config.GlobalServerConfig.Redis.Database,
	})
}
