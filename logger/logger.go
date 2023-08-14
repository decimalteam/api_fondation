package logger

import (
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

func GetLogger(logLevel uint32, logFormat, timestampFormat string) *logrus.Logger {
	return &logrus.Logger{
		Level: logrus.Level(logLevel),
		Formatter: &easy.Formatter{
			TimestampFormat: timestampFormat,
			LogFormat:       logFormat,
		},
	}
}
