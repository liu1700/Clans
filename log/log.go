package log

import (
	"fmt"

	"go.uber.org/zap"
)

// 无需结构化的数据
var sugeredLogger *zap.SugaredLogger

const (
	DEV  = 1
	PROD = 2
)

// 初始化
func InitLogger(logLevel int) {
	logger, _ := zap.NewDevelopment()
	if logLevel == PROD {
		logger, _ = zap.NewProduction()
	}
	sugeredLogger = logger.Sugar()
}

// 各种log
func Info(infos ...interface{}) {
	defer sugeredLogger.Sync()
	sugeredLogger.Info(infos)
}

func Infof(templateStr string, infos ...interface{}) {
	defer sugeredLogger.Sync()
	sugeredLogger.Infof(templateStr, infos)
}

func Warning(infos ...interface{}) {
	defer sugeredLogger.Sync()
	sugeredLogger.Warn(infos)
}

func Warningf(templateStr string, infos ...interface{}) {
	defer sugeredLogger.Sync()
	sugeredLogger.Warnf(templateStr, infos)
}

func Error(infos ...interface{}) {
	defer sugeredLogger.Sync()
	sugeredLogger.Error(infos)
}

func Errorf(templateStr string, infos ...interface{}) {
	defer sugeredLogger.Sync()
	sugeredLogger.Errorf(templateStr, infos)
}

func Println(infos ...interface{}) {
	defer sugeredLogger.Sync()
	fmt.Println(infos)
}
