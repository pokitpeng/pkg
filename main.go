package main

import "github.com/pokitpeng/pkg/logger"

func main() {
	//log := logger.NewDevelopLog()
	//log.Info("hello")

	log := logger.NewLog(logger.Config{
		IsStdOut: true,
		Format:   logger.FormatConsole,
		Encoder:  logger.EncoderCapitalColor,
		Level:    logger.LevelDebug,
	})
	log.Info("hi")
}
