package initialize

import (
	"fmt"

	"github.com/Rinai-R/ApexLecture/server/cmd/push/config"
	"github.com/redis/go-redis/v9"
)

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
