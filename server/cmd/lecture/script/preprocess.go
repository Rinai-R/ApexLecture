package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/config"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/bytedance/sonic"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	// 待会需要用到的变量
	var conf *viper.Viper
	var err error
	var EtcdConf config.EtcdConfig
	var Registry *clientv3.Client
	var ServerConfig config.ServerConfig
	var byteData []byte
	var content *clientv3.GetResponse

	// 读取配置文件
	conf = viper.New()
	conf.SetConfigFile(consts.LectureConfig)
	err = conf.ReadInConfig()
	if err != nil {
		panic("PreProcess failed: ReadInConfig failed" + err.Error())
	}

	// 解析etcd配置
	EtcdConf = config.EtcdConfig{}
	err = conf.Unmarshal(&EtcdConf)
	if err != nil {
		panic("PreProcess failed: Unmarshal EtcdConfig failed" + err.Error())
	}
	Registry, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{net.JoinHostPort(EtcdConf.Host, EtcdConf.Port)},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic("PreProcess failed: New Etcd client failed" + err.Error())
	}

	// 预先准备 ServerConfig 的数据
	ServerConfig = config.ServerConfig{
		Name: consts.LectureSrvPrefix,
		Host: consts.LectureHost,
		Port: consts.LecturePort,
		Mysql: config.MysqlConfig{
			Host:     consts.MysqlHost,
			Port:     consts.MysqlPort,
			Username: consts.MysqlUser,
			Password: consts.MysqlPassword,
			Database: consts.MysqlDatabase,
		},
		Minio: config.MinioConfig{
			Endpoint:        consts.MinioEndpoint,
			AccessKeyID:     consts.MinioAccessKey,
			SecretAccessKey: consts.MinioSecretKey,
			BucketName:      consts.MinioBucket,
			Secure:          consts.MinioSecure,
		},
		Redis: config.RedisConfig{
			Host:     consts.RedisHost,
			Port:     consts.RedisPort,
			Password: consts.RedisPassword,
			Database: consts.RedisDatabase,
		},
		OtelEndpoint: consts.OtelEndpoint,
	}
	// 序列化 ServerConfig
	byteData, err = sonic.Marshal(ServerConfig)
	if err != nil {
		panic("PreProcess failed: json.Marshal failed" + err.Error())
	}
	// 写入etcd
	Registry.Put(context.Background(), EtcdConf.Key, string(byteData))

	// 最后的验证
	content, err = Registry.Get(context.Background(), EtcdConf.Key)
	if err != nil {
		panic("PreProcess failed: Get failed" + err.Error())
	}

	for _, v := range content.Kvs {
		fmt.Println(string(v.Value))
	}
	Registry.Close()
}
