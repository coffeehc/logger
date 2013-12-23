// logger_test.go
package logger

import (
	"fmt"
	"path"
	"runtime"
	"testing"
)

func TestBaseLog(t *testing.T) {
	//Debugf("Test")
	_, fulleFilename, _, ok := runtime.Caller(0)
	if !ok {
		panic("获取Logger根路径出错")
	}
	fmt.Printf("fulleFilename :%s\n", fulleFilename)
	rootPath := path.Base(fulleFilename)
	fmt.Printf("rootPath :%s\n", rootPath)
}
