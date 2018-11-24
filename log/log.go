package log

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()
}

func Debug(format string, a ...interface{}) {
	logger.Debug(fmt.Sprintf(format, a...))
}

func Info(format string, a ...interface{}) {
	logger.Info(fmt.Sprintf(format, a...))
}

func Error(format string, a ...interface{}) {
	logger.Error(fmt.Sprintf(format, a...))
}

func Fatal(format string, a ...interface{}) {
	logger.Fatal(fmt.Sprintf(format, a...))
}
