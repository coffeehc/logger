// logger_test.go
package logger

import (
	"testing"
	"time"
)

func TestBaseLog(t *testing.T) {
	addStdOutFilter(LOGGER_LEVEL_DEBUG, "/", "", "%T-%L-%C-%M")
	//	addFileFilterForTime(LOGGER_LEVEL_DEBUG, "/", "/Users/coffee/logs/testLog.log", time.Second*10, 2)
	//	addFileFilterForDefualt(LOGGER_LEVEL_DEBUG, "/", "/Users/coffee/logs/testLog.log")
	//	addFileFilterForSize(LOGGER_LEVEL_DEBUG, "/", "/Users/coffee/logs/testLog.log", 3*1024, 2)
	for a := 0; a < 10000; a++ {
		Debug("test,time is %s ====%d", time.Now(), a)
	}

}
