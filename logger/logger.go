package logger

import (
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"time"
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

func LogGinMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Starting time request
		startTime := time.Now()

		// Processing request
		ctx.Next()

		// End Time request
		endTime := time.Now()

		// execution time
		latencyTime := endTime.Sub(startTime)

		// Request method
		reqMethod := ctx.Request.Method

		// Request route
		reqUri := ctx.Request.RequestURI

		// status code
		statusCode := ctx.Writer.Status()

		// Request IP
		clientIP := ctx.ClientIP()

		log.WithFields(log.Fields{
			"METHOD":    reqMethod,
			"URI":       reqUri,
			"STATUS":    statusCode,
			"LATENCY":   latencyTime,
			"CLIENT_IP": clientIP,
		}).Info("HTTP REQUEST")

		ctx.Next()
	}
}
