// config

/*
 使用goconfig来获取配置，格式如下：
-
 level: debug
 package_path: /
 adapter: console
-
 level: error
 package_path: /
 adapter: file
 log_path: /logs/box/box.log
 rotate: 3
 #备份策略：size or time  or default
 rotate_policy: time
 #备份范围：如果策略是time则表示时间间隔N分钟，如果是size则表示每个日志的最大大小(MB)
 rotate_scope: 10
*/

package logger

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Logger_config struct {
	Level         string //日志级别
	Package_path  string //日志路径
	Adapter       string //适配器,console,file两种
	Rotate        int    //日志切割个数
	Rotate_policy string //切割策略,time or size or  default
	Rotate_scope  int64  //切割范围:如果按时间切割则表示的n分钟,如果是size这表示的是文件大小MB
	Log_path      string //如果适配器使用的file则用来指定文件路径
}

const (
	CONFIG_SELECT        = "log"
	CONFIG_SELECT_PREFIX = "log_"
	CONFIG_LEVEL         = "level"
	CONFIG_PATH          = "path"
	CONFIG_ADAPTER       = "adapter"
	CONFIG_LOGGERS       = "loggers"
	CONFIG_ROTATE        = "rotate"
	CONFIG_ROTATEPOLICY  = "rotatePolicy"
	CONFIG_ROTATESCOPE   = "rotateScope"
	CONFIG_LOGPATH       = "logPath"
)

var _loggerConf *string = flag.String("loggerConf", "", "日志文件路径")

//加载日志配置,如果指定了-loggerConf参数,则加载这个参数指定的配置文件,如果没有则使用默认的配置
func loadLoggerConfig() {
	if !flag.Parsed() {
		flag.Parse()
	}
	confs := parseConfile(*_loggerConf)
	if confs == nil {
		confs = []*Logger_config{&Logger_config{Level: "debug", Package_path: "/", Adapter: "console"}}
	}
	for _, conf := range confs {
		addLogger(conf)
	}
}

func parseConfile(loggerConf string) (confs []*Logger_config) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("解析配置文件%s出错:%s\n", loggerConf, r)
		}
	}()
	if *_loggerConf != "" {
		data, err := ioutil.ReadFile(loggerConf)
		if err != nil {
			fmt.Printf("加载文件%s错误:%s\n", loggerConf, err)
		} else {
			yaml.Unmarshal(data, confs)
		}
	}
	return
}

//添加日志配置
func addLogger(conf *Logger_config) {
	switch conf.Adapter {
	case "file":
		addFileLogger(conf)
		break
	case "console":
		addConsoleLogger(conf)
		break
	default:
		fmt.Printf("不能识别的日志适配器:%s", conf.Adapter)
	}
}

//添加console的日志配置
func addConsoleLogger(conf *Logger_config) {
	addStdOutFilter(getLevel(conf.Level), conf.Package_path, LOGGER_TIMEFORMAT_NANOSECOND)
}

//添加文件系统的日志配置
func addFileLogger(conf *Logger_config) {
	rotatePolicy := strings.ToLower(conf.Rotate_policy)
	switch rotatePolicy {
	case "time":
		addFileFilterForTime(getLevel(conf.Level), conf.Package_path, conf.Log_path, time.Minute*time.Duration(conf.Rotate_scope), conf.Rotate)
		return
	case "size":
		addFileFilterForSize(getLevel(conf.Level), conf.Package_path, conf.Log_path, conf.Rotate_scope*1048576, conf.Rotate)
		return
	default:
		addFileFilterForDefualt(getLevel(conf.Level), conf.Package_path, conf.Log_path)
		return
	}
}

func getLevel(level string) byte {
	level = strings.ToLower(level)
	switch level {
	case "trace":
		return LOGGER_LEVEL_TRACE
	case "debug":
		return LOGGER_LEVEL_DEBUG
	case "info":
		return LOGGER_LEVEL_INFO
	case "warn":
		return LOGGER_LEVEL_WARN
	case "error":
		return LOGGER_LEVEL_ERROR
	default:
		return LOGGER_LEVEL_DEBUG
	}
}
