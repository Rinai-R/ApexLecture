package initialize

import (
	"context"
	"time"

	"github.com/Rinai-R/ApexLecture/server/cmd/api/config"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func InitConfig() {
	// 需要用到的变量
	var conf *viper.Viper
	var err error
	var Registry *clientv3.Client
	var content *clientv3.GetResponse
	var byteData []byte

	// 初始化配置
	conf = viper.New()
	conf.SetConfigFile(consts.ApiConfig)
	err = conf.ReadInConfig()
	if err != nil {
		hlog.Fatal("initialize: Error reading config file:", err)
	}
	err = conf.Unmarshal(&config.GlobalEtcdConfig)
	if err != nil {
		hlog.Fatal("initialize: Error unmarshalling etcd config:", err)
	}

	// 初始化etcd
	Registry, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{config.GlobalEtcdConfig.Host + ":" + config.GlobalEtcdConfig.Port},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		hlog.Fatal("initialize: Error connecting to etcd: ", err)
	}

	// 从etcd获取配置
	content, err = Registry.Get(context.Background(), config.GlobalEtcdConfig.Key)
	if err != nil {
		hlog.Fatal("initialize: Error getting config from etcd: ", err)
	}
	byteData = []byte(content.Kvs[0].Value)
	err = sonic.Unmarshal(byteData, &config.GlobalServerConfig)
	if err != nil {
		hlog.Fatal("initialize: Error unmarshalling server config:", err)
	}

	// 关闭etcd连接
	Registry.Close()
	hlog.Info("initialize: Config initialized successfully")
}
