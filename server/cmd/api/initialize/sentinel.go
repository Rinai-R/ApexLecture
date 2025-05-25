package initialize

import (
	"github.com/Rinai-R/ApexLecture/server/cmd/api/config"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func InitSentinel() {
	err := sentinel.InitDefault()
	if err != nil {
		hlog.Fatal("Initialize: init sentinel error ", err)
	}

	_, err = flow.LoadRules([]*flow.Rule{
		{
			Resource:               config.GlobalServerConfig.Name,
			Threshold:              1000,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
		},
	})
}
