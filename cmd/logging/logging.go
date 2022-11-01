package logging

import (
	"log"

	"go.uber.org/zap"
)

func InitLogger() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("cannot init logger ", err)
	}

	return logger
}
