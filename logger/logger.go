package logger

import (
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

func GetLogger(logLevel uint32, fieldsOrder []string, timestampFormat string, hideKeys, noColors bool) *logrus.Logger {
	return &logrus.Logger{
		Level: logrus.Level(logLevel),
		Formatter: &nested.Formatter{
			FieldsOrder:     fieldsOrder,
			TimestampFormat: timestampFormat,
			HideKeys:        hideKeys,
			NoColors:        noColors,
		},
	}
}
