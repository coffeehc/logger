// logger project logger.go
package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
	"utils"
)

const (
	LOGGER_LEVEL_ERROR byte = 1 << 0
	LOGGER_LEVEL_WARN  byte = 1<<1 | LOGGER_LEVEL_ERROR
	LOGGER_LEVEL_INFO  byte = 1<<2 | LOGGER_LEVEL_WARN
	LOGGER_LEVEL_DEBUG byte = 1<<3 | LOGGER_LEVEL_INFO
	LOGGER_LEVEL_TRACE byte = 1<<4 | LOGGER_LEVEL_DEBUG

	LOGGER_DEFAULT_BUFSIZE int           = 1024
	LOGGER_DEFAULT_TIMEOUT time.Duration = time.Second * 10

	LOGGER_TIMEFORMAT_SECOND     string = "2006-01-02 15:04:05"
	LOGGER_TIMEFORMAT_NANOSECOND string = "2006-01-02 15:04:05.999999999"
	LOGGER_TIMEFORMAT_ALL        string = "2006-01-02 15:04:05.999999999 -0700 UTC"
)

func getLevelStr(level byte) string {
	switch level {
	case LOGGER_LEVEL_ERROR:
		return "error"
	case LOGGER_LEVEL_WARN:
		return "warn"
	case LOGGER_LEVEL_INFO:
		return "info"
	case LOGGER_LEVEL_DEBUG:
		return "debug"
	case LOGGER_LEVEL_TRACE:
		return "trace"
	default:
		return "--"
	}
}

type Flusher interface {
	Flush()
}
type logFilter struct {
	level      byte   //拦截级别
	path       string //拦截路径
	timeFormat string //时间戳格式
	out        io.Writer
	cache      chan string
}

func (this *logFilter) canSave(level byte, lineInfo string) bool {
	return this.level&level == level && strings.HasPrefix(lineInfo, this.path)
}

func (this *logFilter) run() {
	for {
		select {
		case content := <-this.cache:
			this.out.Write([]byte(content))
			continue
		case <-time.After(time.Second * 5):
			if v, ok := this.out.(Flusher); ok {
				v.Flush()
			}
			continue
		}
	}
}

var evnRootPathLen int
var isAddFilter bool = false
var filters map[string]*logFilter

func init() {
	filters = make(map[string]*logFilter)
	rootPath, err := os.Getwd()
	if err != nil {
		fmt.Println("初始化日志出现一个错误:%v", err)
	}
	evnRootPathLen = len(rootPath)
	AddStdOutFilter("ROOT", LOGGER_LEVEL_DEBUG, "/", "")
	Info("logger框架初始化完成")
}

func AddStdOutFilter(name string, level byte, path string, timeFormat string) {
	if timeFormat == "" {
		timeFormat = LOGGER_TIMEFORMAT_NANOSECOND
	}
	if !isAddFilter {
		delete(filters, "ROOT")
	}
	AddFileter(name, level, path, timeFormat, os.Stdout)
	isAddFilter = true
}

func AddFileter(name string, level byte, path string, timeFormat string, out io.Writer) {
	defer func() {
		if x := recover(); x != nil {
			output(LOGGER_LEVEL_ERROR, fmt.Sprint(x))
		}
	}()
	if path == "" {
		panic("拦截器名称不能为空")
	}
	if timeFormat == "" {
		timeFormat = LOGGER_TIMEFORMAT_SECOND
	}
	if out == nil {
		panic("拦截器名称不能为空")
	}
	if filters[name] != nil {
		panic(fmt.Sprintf("已经定义了一个%s的日志拦截器", name))
	}
	filter := new(logFilter)
	filter.level = level
	filter.path = path
	filter.timeFormat = timeFormat
	filter.out = out
	filter.cache = make(chan string, 200)
	filters[name] = filter
	go filter.run()

}
func output(level byte, content string) {
	_, file, line, ok := runtime.Caller(2)
	var lineInfo string = "-:0"
	if ok {
		file = utils.SubString(file, evnRootPathLen, 1000)
		if !strings.HasPrefix(file, "/") {
			file = "/" + file
		}
		lineInfo = file + ":" + strconv.Itoa(line)
	}
	content = fmt.Sprintf("\t%s\t%s\t%s\n", getLevelStr(level), lineInfo, content)
	for _, filter := range filters {
		if filter != nil && filter.canSave(level, file) {
			filter.cache <- time.Now().Format(filter.timeFormat) + content
		}
	}
}

func Trace(v ...interface{}) {
	output(LOGGER_LEVEL_TRACE, fmt.Sprint(v...))
}

func Tracef(format string, v ...interface{}) {
	output(LOGGER_LEVEL_TRACE, fmt.Sprintf(format, v...))
}

func Debug(v ...interface{}) {
	output(LOGGER_LEVEL_DEBUG, fmt.Sprint(v...))
}

func Debugf(format string, v ...interface{}) {
	output(LOGGER_LEVEL_DEBUG, fmt.Sprintf(format, v...))
}

func Info(v ...interface{}) {
	output(LOGGER_LEVEL_INFO, fmt.Sprint(v...))
}

func Infof(format string, v ...interface{}) {
	output(LOGGER_LEVEL_INFO, fmt.Sprintf(format, v...))
}

func Warn(v ...interface{}) {
	output(LOGGER_LEVEL_WARN, fmt.Sprint(v...))
}

func Warnf(format string, v ...interface{}) {
	output(LOGGER_LEVEL_WARN, fmt.Sprintf(format, v...))
}

func Error(v ...interface{}) {
	output(LOGGER_LEVEL_ERROR, fmt.Sprint(v...))
}

func Errorf(format string, v ...interface{}) {
	output(LOGGER_LEVEL_ERROR, fmt.Sprintf(format, v...))
}
