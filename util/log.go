package util

import (
	"go.uber.org/zap"
)

var Log *zap.SugaredLogger

func init() {
	var cfg = zap.NewDevelopmentConfig()

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	Log = logger.Sugar()
	//Log.Infof("test")
}
