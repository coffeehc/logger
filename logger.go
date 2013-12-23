// logger project logger.go
package logger

import (
	"fmt"
	"github.com/coffeehc/utils"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
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

var (
	isAddFilter    bool = false
	filters        map[string]*logFilter
	isInit         bool = false
	evnRootPathLen int
)

func StartDevModel() {
	Init()
	AddStdOutFilter("ROOT", LOGGER_LEVEL_DEBUG, "/", "")
}

func Init() {
	if isInit {
		Info("日志已经初始化,不需要再次初始化")
		return
	}
	_, fulleFilename, _, ok := runtime.Caller(0)
	if !ok {
		panic("获取Logger根路径出错")
	}
	evnRootPathLen = strings.Index(fulleFilename, "/src/") + 4
	filters = make(map[string]*logFilter)
	//AddStdOutFilter("ROOT", LOGGER_LEVEL_DEBUG, "/", "")
	isAddFilter = false
	Info("logger框架初始化完成")
	isInit = true
}

func AddStdOutFilter(name string, level byte, path string, timeFormat string) {
	if timeFormat == "" {
		timeFormat = LOGGER_TIMEFORMAT_NANOSECOND
	}
	AddFileter(name, level, path, timeFormat, os.Stdout)
}

func AddFileter(name string, level byte, path string, timeFormat string, out io.Writer) {
	if !isAddFilter {
		delete(filters, "ROOT")
	}
	isAddFilter = true
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
	Debugf("添加了Log Filter:%s,%v,%s", name, level, path)
}
func output(level byte, content string) {
	_, file, line, ok := runtime.Caller(2)
	var lineInfo string = "-:0"
	if ok {
		file = utils.SubString(file, evnRootPathLen, 1000)
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
