package main

import (
	"context"
	"log"

	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/pkg/goroutine"
	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/config"
	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/initialize"
	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/mq"
	service "github.com/Rinai-R/ApexLecture/server/cmd/quiz/pkg/quiz_status"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/quiz/quizservice"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/obs-opentelemetry/provider"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
)

func main() {

	initialize.Initlogger()
	initialize.InitConfig()
	d := initialize.InitDB()
	rdb := initialize.InitRedis()
	pro, con := initialize.InitMQ()
	r, i := initialize.InitRegistry()
	handler := mq.NewConsumerHandler(dao.NewMysqlManager(d))
	p := provider.NewOpenTelemetryProvider(
		provider.WithServiceName(config.GlobalServerConfig.Name),
		provider.WithExportEndpoint(config.GlobalServerConfig.OtelEndpoint),
		provider.WithInsecure(),
	)
	defer p.Shutdown(context.Background())

	go func() {
		consumer := mq.NewConsumerManager(con)
		err := consumer.Consume(context.Background(), config.GlobalServerConfig.Kafka.Topic, handler)
		if err != nil {
			klog.Error("Consume failed", err)
		}
	}()
	svr := quizservice.NewServer(
		&QuizServiceImpl{
			MysqlManager:      dao.NewMysqlManager(d),
			RedisManager:      dao.NewRedisManager(rdb),
			ProducerManager:   mq.NewProducerManager(pro),
			goroutinePool:     goroutine.NewPool(1000),
			QuizStatusHanlder: service.NewQuizStatusHanlder(dao.NewRedisManager(rdb)),
		},
		server.WithRegistry(r),
		server.WithRegistryInfo(i),
		server.WithServiceAddr(i.Addr),
		server.WithSuite(tracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: config.GlobalServerConfig.Name,
		}))
	err := svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}
