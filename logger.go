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
)

type Logger interface {
	Trace(format string, v ...interface{})
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}

func GetLogger() Logger {
	return loggercopy
}

var loggercopy _logger

type _logger struct {
}

func (this _logger) Trace(format string, v ...interface{}) {
	output(LOGGER_LEVEL_TRACE, fmt.Sprintf(format, v...), 3)
}

func (this _logger) Debug(format string, v ...interface{}) {
	output(LOGGER_LEVEL_DEBUG, fmt.Sprintf(format, v...), 3)
}

func (this _logger) Info(format string, v ...interface{}) {
	output(LOGGER_LEVEL_INFO, fmt.Sprintf(format, v...), 3)
}

func (this _logger) Warn(format string, v ...interface{}) {
	output(LOGGER_LEVEL_WARN, fmt.Sprintf(format, v...), 3)
}

func (this _logger) Error(format string, v ...interface{}) {
	output(LOGGER_LEVEL_ERROR, fmt.Sprintf(format, v...), 3)
}

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
	code_Level                   int    = 2
)

func getLevelStr(level byte) string {
	switch level {
	case LOGGER_LEVEL_ERROR:
		return "ERROR"
	case LOGGER_LEVEL_WARN:
		return "WARN"
	case LOGGER_LEVEL_INFO:
		return "INFO"
	case LOGGER_LEVEL_DEBUG:
		return "DEBUG"
	case LOGGER_LEVEL_TRACE:
		return "TRACE"
	default:
		return "--"
	}
}

type Flusher interface {
	Flush() error
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

var (
	filters        map[string]*logFilter
	evnRootPathLen int
)

func init() {
	filters = make(map[string]*logFilter)
	AddStdOutFilter("_ROOT", LOGGER_LEVEL_INFO, "/", "")
	Info("logger框架初始化完成")
}

func StartDevModul(level byte) {
	AddStdOutFilter("_ROOT", level, "/", "")
}

func AddStdOutFilter(name string, level byte, path string, timeFormat string) {
	if timeFormat == "" {
		timeFormat = LOGGER_TIMEFORMAT_NANOSECOND
	}
	AddFileter(name, level, path, timeFormat, os.Stdout)
}

func AddFileter(name string, level byte, path string, timeFormat string, out io.Writer) {
	delete(filters, "_ROOT")
	defer func() {
		if x := recover(); x != nil {
			output(LOGGER_LEVEL_ERROR, fmt.Sprint(x), code_Level)
		}
	}()
	if name == "" {
		panic("拦截器名称不能为空")
	}
	if filters[name] != nil {
		Warn("已经定义了一个%s的日志拦截器,不能再次定义", name)
		return
	}
	if path == "" {
		panic("拦截器拦截路径不能为空")
	}
	if timeFormat == "" {
		timeFormat = LOGGER_TIMEFORMAT_SECOND
	}
	if out == nil {
		panic("拦截器输出不能为空")
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
func output(logLevel byte, content string, codeLevel int) {
	_, file, line, ok := runtime.Caller(codeLevel)
	var lineInfo string = "-:0"
	if ok {
		index := strings.Index(file, "/src/") + 4
		lineInfo = file[index:] + ":" + strconv.Itoa(line)
	}
	content = fmt.Sprintf("\t%s\t%s\t%s\n", getLevelStr(logLevel), lineInfo, content)
	for _, filter := range filters {
		if filter != nil && filter.canSave(logLevel, lineInfo) {
			filter.cache <- time.Now().Format(filter.timeFormat) + content
		}
	}
}

func Trace(format string, v ...interface{}) {
	output(LOGGER_LEVEL_TRACE, fmt.Sprintf(format, v...), code_Level)
}

func Debug(format string, v ...interface{}) {
	output(LOGGER_LEVEL_DEBUG, fmt.Sprintf(format, v...), code_Level)
}

func Info(format string, v ...interface{}) {
	output(LOGGER_LEVEL_INFO, fmt.Sprintf(format, v...), code_Level)
}

func Warn(format string, v ...interface{}) {
	output(LOGGER_LEVEL_WARN, fmt.Sprintf(format, v...), code_Level)
}

func Error(format string, v ...interface{}) {
	output(LOGGER_LEVEL_ERROR, fmt.Sprintf(format, v...), code_Level)
}
