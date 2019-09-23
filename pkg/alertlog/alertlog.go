package alertlog

import (
	"go.uber.org/zap"
	"log"
)


var (
	Logger = initLogger()
)

func initLogger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	//logger, err := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	return logger.Sugar()
}

//sugar.Infow("failed to fetch URL",
//// Structured context as loosely typed key-value pairs.
//"url", url,
//"attempt", 3,
//"backoff", time.Second,
//)
//sugar.Infof("Failed to fetch URL: %s", url)