package initialize

import (
	"io"
	"os"
	"path"
	"time"

	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/kitex-contrib/obs-opentelemetry/logging/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Initlogger() {
	logFilePath := consts.KlogFilePath
	if err := os.MkdirAll(logFilePath, 0o777); err != nil {
		panic(err)
	}
	LogFileName := time.Now().Format("2006-01-02 15:04:05") + ".log"
	fileName := path.Join(logFilePath, LogFileName)
	if _, err := os.Stat(logFilePath + LogFileName); err != nil {
		_, err := os.Create(fileName)
		if err != nil {
			panic("initialize: create log file error " + err.Error())
		}
	}
	logger := logrus.NewLogger()
	lumberjackLogger := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    20,
		MaxBackups: 5,
		MaxAge:     10,
		Compress:   true,
	}
	fileWriter := io.MultiWriter(os.Stdout, lumberjackLogger)
	logger.SetOutput(fileWriter)
	logger.SetLevel(klog.LevelInfo)
	klog.SetLogger(logger)
	klog.Info("initialize: initialize logger success")
}
