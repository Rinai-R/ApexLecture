package initialize

import (
	"fmt"

	"github.com/Rinai-R/ApexLecture/server/cmd/chat/config"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/streadway/amqp"
)

func InitMqConn() *amqp.Connection {
	conn, err := amqp.Dial(
		fmt.Sprintf(consts.RabbitMqDNS,
			config.GlobalServerConfig.RabbitMQ.Username,
			config.GlobalServerConfig.RabbitMQ.Password,
			config.GlobalServerConfig.RabbitMQ.Host,
			config.GlobalServerConfig.RabbitMQ.Port,
		),
	)
	if err != nil {
		klog.Fatal("Failed to connect to RabbitMQ ", err)
	}
	return conn
}
