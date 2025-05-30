package initialize

import (
	"io"
	"os"
	"path"
	"time"

	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/obs-opentelemetry/logging/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Initlogger() {
	logFilePath := consts.HlogFilePath
	if err := os.MkdirAll(logFilePath, 0o777); err != nil {
		panic("initialize: mkdir log failed")
	}
	logFileName := time.Now().Format("2006-01-02 15:04:05") + ".log"
	fileName := path.Join(logFilePath, logFileName)
	if _, err := os.Stat(fileName); err == nil {
		if _, err = os.Create(fileName); err != nil {
			panic("initialize: create log failed")
		}
	}
	logger := logrus.NewLogger()
	// lumberjack 可以帮助自动管理日志
	lumberjackLogger := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    20,
		MaxBackups: 5,
		MaxAge:     10,
		Compress:   true,
	}
	fileWriter := io.MultiWriter(os.Stdout, lumberjackLogger)
	logger.SetOutput(fileWriter)
	logger.SetLevel(hlog.LevelInfo)
	hlog.SetLogger(logger)
	hlog.Info("initialize: logger initialized")
}
