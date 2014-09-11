// logger_test.go
package logger

import (
	"testing"
	"time"
)

func TestBaseLog(t *testing.T) {
	//AddFileFilterForTime("testTimeLog", LOGGER_LEVEL_DEBUG, "/", "d:/testlog/testLog.log", time.Second*10, 10)
	//AddFileFilterForDefualt("testTimeLog", LOGGER_LEVEL_DEBUG, "/", "d:/testlog/testLog.log")
	//AddFileFilterForSize("testTimeLog", LOGGER_LEVEL_DEBUG, "/", "d:/testlog/testLog.log", 3*1024, 10)
	for {
		Debugf("test,time is %s", time.Now())
		time.Sleep(time.Millisecond * 3)
	}
}
