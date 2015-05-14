// logger project logger.go
package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	FORMAT_TIME     = "%T"
	FORMAT_LEVEL    = "%L"
	FORMAT_CODEINFO = "%C"
	FORMAT_MESSGAE  = "%M"
)

type Logger interface {
	Trace(format string, v ...interface{}) string
	Debug(format string, v ...interface{}) string
	Info(format string, v ...interface{}) string
	Warn(format string, v ...interface{}) string
	Error(format string, v ...interface{}) string
}

func GetLogger() Logger {
	return loggercopy
}

var loggercopy _logger

type _logger struct {
}

func (this _logger) Trace(format string, v ...interface{}) string {
	return output(LOGGER_LEVEL_TRACE, fmt.Sprintf(format, v...), 3)
}

func (this _logger) Debug(format string, v ...interface{}) string {
	return output(LOGGER_LEVEL_DEBUG, fmt.Sprintf(format, v...), 3)
}

func (this _logger) Info(format string, v ...interface{}) string {
	return output(LOGGER_LEVEL_INFO, fmt.Sprintf(format, v...), 3)
}

func (this _logger) Warn(format string, v ...interface{}) string {
	return output(LOGGER_LEVEL_WARN, fmt.Sprintf(format, v...), 3)
}

func (this _logger) Error(format string, v ...interface{}) string {
	return output(LOGGER_LEVEL_ERROR, fmt.Sprintf(format, v...), 3)
}

const (
	LOGGER_LEVEL_ERROR byte = 1 << 0
	LOGGER_LEVEL_WARN  byte = 1<<1 | LOGGER_LEVEL_ERROR
	LOGGER_LEVEL_INFO  byte = 1<<2 | LOGGER_LEVEL_WARN
	LOGGER_LEVEL_DEBUG byte = 1<<3 | LOGGER_LEVEL_INFO
	LOGGER_LEVEL_TRACE byte = 1<<4 | LOGGER_LEVEL_DEBUG

	LOGGER_DEFAULT_BUFSIZE int           = 1024
	LOGGER_DEFAULT_TIMEOUT time.Duration = time.Second * 1

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

//日志持久化接口
type Flusher interface {
	Flush() error
}

//日志拦截器定义
type logFilter struct {
	level      byte        //拦截级别
	path       string      //拦截路径
	timeFormat string      //时间戳格式
	format     []logFormat //日志格式
	out        io.Writer
	cache      chan *logContent
	stop       chan bool
}

//判断是否需要过滤器处理
func (this *logFilter) canSave(level byte, lineInfo string) bool {
	return this.level&level == level && strings.HasPrefix(lineInfo, this.path)
}

//过滤器后台输出goruntine
func (this *logFilter) run() {
	buf := bytes.NewBuffer(nil)
	for {
		select {
		case content := <-this.cache:
			if this.canSave(content.level, content.lineInfo) {
				buf.Reset()
				for _, f := range this.format {
					buf.Write(f.data)
					switch f.appendContent {
					case FORMAT_TIME:
						buf.WriteString(content.appendTime.Format(this.timeFormat))
					case FORMAT_LEVEL:
						buf.WriteString(getLevelStr(content.level))
					case FORMAT_CODEINFO:
						buf.WriteString(content.lineInfo)
					case FORMAT_MESSGAE:
						buf.WriteString(content.content)
					default:
					}
				}
				buf.WriteByte('\n')
				buf.WriteTo(this.out)
			}
			continue
		case <-time.After(time.Millisecond * 500):
			if v, ok := this.out.(Flusher); ok {
				v.Flush()
			}
			continue
		case <-this.stop:
			return
		}
	}
}

func (this *logFilter) clear() {
	this.stop <- true
}

var (
	filters        []*logFilter
	evnRootPathLen int
)

func init() {
	loadLoggerConfig(*_loggerConf)
}

func addStdOutFilter(level byte, path string, timeFormat string, format string) {
	AddFileter(level, path, timeFormat, format, os.Stdout)
}

//清空过滤器,主要用于自定义处理日志
func ClearFilter() {
	filters = make([]*logFilter, 0)
}

//添加日志过滤器,参数说明:级别,包路径,时间格式,Writer接口
func AddFileter(level byte, path string, timeFormat string, format string, out io.Writer) {
	if timeFormat == "" {
		timeFormat = LOGGER_TIMEFORMAT_SECOND
	}
	defer func() {
		if x := recover(); x != nil {
			output(LOGGER_LEVEL_ERROR, fmt.Sprint(x), code_Level)
		}
	}()
	if path == "" {
		panic("拦截器拦截路径不能为空")
	}
	if timeFormat == "" {
		timeFormat = LOGGER_TIMEFORMAT_SECOND
	}
	if format == "" {
		format = "%T %L %C %M"
	}
	if out == nil {
		panic("拦截器输出不能为空")
	}
	filter := new(logFilter)
	filter.level = level
	filter.path = path
	filter.timeFormat = timeFormat
	filter.out = out
	filter.format = parseFormat(format)
	filter.cache = make(chan *logContent, 200)
	filter.stop = make(chan bool, 1)
	filters = append(filters, filter)
	go filter.run()
}

type logFormat struct {
	data          []byte
	appendContent string
}

func parseFormat(format string) []logFormat {
	fs := make([]logFormat, 0)
	fc := []byte(format)
	start := 0
	i := 0
	for ; i < len(fc); i++ {
		b := fc[i]
		if b == '%' {
			switch fc[i+1] {
			case 'T':
				fs = append(fs, logFormat{data: fc[start:i], appendContent: FORMAT_TIME})
				i += 1
				start = i + 1
				continue
			case 'L':
				fs = append(fs, logFormat{data: fc[start:i], appendContent: FORMAT_LEVEL})
				i += 1
				start = i + 1
				break
			case 'C':
				fs = append(fs, logFormat{data: fc[start:i], appendContent: FORMAT_CODEINFO})
				i += 1
				start = i + 1
				break
			case 'M':
				fs = append(fs, logFormat{data: fc[start:i], appendContent: FORMAT_MESSGAE})
				i += 1
				start = i + 1
				break
			}
		}
	}
	if i > start {
		fs = append(fs, logFormat{data: fc[start : i-1]})
	}
	return fs
}

type logContent struct {
	appendTime time.Time
	level      byte
	lineInfo   string
	content    string
}

// real out implement
func output(logLevel byte, content string, codeLevel int) string {
	_, file, line, ok := runtime.Caller(codeLevel)
	var lineInfo string = "-:0"
	if ok {
		index := strings.Index(file, "/src/") + 4
		lineInfo = file[index:] + ":" + strconv.Itoa(line)
	}
	log := &logContent{time.Now(), logLevel, lineInfo, content}
	for _, filter := range filters {
		if filter != nil {
			filter.cache <- log
		}
	}
	return content
}

//print trace log
func Trace(format string, v ...interface{}) string {
	return output(LOGGER_LEVEL_TRACE, fmt.Sprintf(format, v...), code_Level)
}

//print debug log
func Debug(format string, v ...interface{}) string {
	return output(LOGGER_LEVEL_DEBUG, fmt.Sprintf(format, v...), code_Level)
}

//print Info log
func Info(format string, v ...interface{}) string {
	return output(LOGGER_LEVEL_INFO, fmt.Sprintf(format, v...), code_Level)
}

//print Warn log
func Warn(format string, v ...interface{}) string {
	return output(LOGGER_LEVEL_WARN, fmt.Sprintf(format, v...), code_Level)
}

// print Error log
func Error(format string, v ...interface{}) string {
	return output(LOGGER_LEVEL_ERROR, fmt.Sprintf(format, v...), code_Level)
}
