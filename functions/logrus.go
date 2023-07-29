package functions

import (
	graylog "github.com/gemnasium/logrus-graylog-hook"
)

func initLogData() *graylog.GraylogHook {
	hook := graylog.NewAsyncGraylogHook("<graylog_ip>:<graylog_port>", map[string]interface{}{"this": "is logged every time"})
	// NOTE: you must call Flush() before your program exits to ensure ALL of your logs are sent.
	// This defer statement will not have that effect if you write it in a non-main() method.
	defer hook.Flush()
	return hook
}

func writeLog() {

}
