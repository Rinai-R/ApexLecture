package middleware

import (
	"context"
	"net/http"

	"github.com/Rinai-R/ApexLecture/server/cmd/api/config"
	"github.com/Rinai-R/ApexLecture/server/shared/rsp"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/cloudwego/hertz/pkg/app"
)

func Sentinel(c context.Context, ctx *app.RequestContext) {
	a, err := sentinel.Entry(
		config.GlobalServerConfig.Name,
		sentinel.WithResourceType(base.ResTypeWeb),
		sentinel.WithTrafficType(base.Inbound))
	if err != nil {
		ctx.JSON(http.StatusTooManyRequests, rsp.ErrorServerBusy())
		ctx.Abort()
		return
	}
	defer a.Exit()
	ctx.Next(c)
}
