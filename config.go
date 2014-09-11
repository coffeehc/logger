// config

/*
	使用goconfig来获取配置，格式如下：

	[log]
	#日志级别:trace,debug,info,warn,error
	level=debug
	path=/
	#日志适配器:console;file
	adapter=console
	#其他日志处理器
	loggers=file

	[log_file]
	level=error
	path=/
	adapter=file
	logPath=/logs/box/box.log
	rotate=3
	#备份策略：size or time  or default
	rotatePolicy=time
	#备份范围：如果策略是time则表示时间间隔N分钟，如果是size则表示每个日志的最大大小(MB)
	rotateScope=
*/

package logger

import (
	"strings"
	"time"

	"github.com/msbranco/goconfig"
)

type logger_config struct {
	name         string
	level        string
	path         string
	adapter      string
	rotate       int
	rotatePolicy string
	rotateScope  int64
	logPath      string
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

func LoadLoggerConfig(config *goconfig.ConfigFile) {
	rootConfig := buildLogConfig(config, CONFIG_SELECT)
	rootConfig.name = "root"
	addLogger(rootConfig)
	loggers, _ := config.GetString(CONFIG_SELECT, CONFIG_LOGGERS)
	if loggers != "" {
		childLoggers := strings.Split(loggers, ",")
		for _, child := range childLoggers {
			if child == "" {
				continue
			}
			childSelect := CONFIG_SELECT_PREFIX + child
			if config.HasSection(childSelect) {
				childConfig := buildLogConfig(config, childSelect)
				childConfig.name = child
				addLogger(childConfig)
			} else {
				Warn("没有找到%s的日志配置", childSelect)
			}
		}
	}
}

func buildLogConfig(config *goconfig.ConfigFile, selectName string) *logger_config {
	level, _ := config.GetString(selectName, CONFIG_LEVEL)
	path, _ := config.GetString(selectName, CONFIG_PATH)
	adapter, _ := config.GetString(selectName, CONFIG_ADAPTER)
	rotate, _ := config.GetInt64(selectName, CONFIG_ROTATE)
	rotatePolicy, _ := config.GetString(selectName, CONFIG_ROTATEPOLICY)
	rotateScope, _ := config.GetInt64(selectName, CONFIG_ROTATESCOPE)
	logPath, _ := config.GetString(selectName, CONFIG_LOGPATH)
	return &logger_config{level: level, path: path, adapter: adapter, rotate: int(rotate), rotatePolicy: rotatePolicy, rotateScope: rotateScope, logPath: logPath}
}

func addLogger(conf *logger_config) {
	if strings.EqualFold("file", conf.adapter) {
		addFileLogger(conf)
	} else {
		addConsoleLogger(conf)
	}
}

func addConsoleLogger(conf *logger_config) {
	AddStdOutFilter(conf.name, getLevel(conf.level), conf.path, LOGGER_TIMEFORMAT_NANOSECOND)
}

func addFileLogger(conf *logger_config) {
	rotatePolicy := strings.ToLower(conf.rotatePolicy)
	switch rotatePolicy {
	case "time":
		AddFileFilterForTime(conf.name, getLevel(conf.level), conf.path, conf.logPath, time.Minute*time.Duration(conf.rotateScope), conf.rotate)
		return
	case "size":
		AddFileFilterForSize(conf.name, getLevel(conf.level), conf.path, conf.logPath, conf.rotateScope*1048576, conf.rotate)
		return
	default:
		AddFileFilterForDefualt(conf.name, getLevel(conf.level), conf.path, conf.logPath)
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
