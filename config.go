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
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type LoggerAppender struct {
	Level         string `yaml:"level"`         //日志级别
	Package_path  string `yaml:"package_path"`  //日志路径
	Adapter       string `yaml:"adapter"`       //适配器,console,file两种
	Rotate        int    `yaml:"rotate"`        //日志切割个数
	Rotate_policy string `yaml:"rotate_policy"` //切割策略,time or size or  default
	Rotate_scope  int64  `yaml:"rotate_scope"`  //切割范围:如果按时间切割则表示的n分钟,如果是size这表示的是文件大小MB
	Log_path      string `yaml:"log_path"`      //如果适配器使用的file则用来指定文件路径
	Timeformat    string `yaml:"timeformat"`    //日志格式
	Format        string `yaml:"format"`
}

type LoggerConfig struct {
	Context   string           `yaml:"context"`
	Appenders []LoggerAppender `yaml:"appenders"`
}

const (
	Adapter_Console = "console"
	Adapter_File    = "file"
)

var _loggerConf *string = flag.String("logger", getDefaultLog(), "日志文件路径")

func getDefaultLog() string {
	file, _ := exec.LookPath(os.Args[0])
	filePath, _ := filepath.Abs(file)
	return path.Join(filepath.Dir(filePath), "conf/log.yml")
}

//加载日志配置,如果指定了-loggerConf参数,则加载这个参数指定的配置文件,如果没有则使用默认的配置
func loadLoggerConfig(loggerConf string) {
	if len(filters) > 0 {
		for _, filter := range filters {
			filter.clear()
		}
	}
	filters = make([]*logFilter, 0)
	conf := parseConfile(loggerConf)
	if conf == nil || len(conf.Appenders) == 0 {
		fmt.Println("没有指定配置文件或者日志配置出错,使用默认配置")
		conf = &LoggerConfig{Context: "Default", Appenders: []LoggerAppender{LoggerAppender{Level: "debug", Package_path: "/", Adapter: "console"}}}
	}
	for _, appender := range conf.Appenders {
		AddAppender(appender)
	}

}

func parseConfile(loggerConf string) *LoggerConfig {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("解析配置文件%s出错:%s\n", loggerConf, r)
		}
	}()
	//log.Printf("加载日志配置文件:%s\n", loggerConf)
	data, err := ioutil.ReadFile(loggerConf)
	if err != nil {
		//log.Printf("[警告]加载日志配置文件错误:%s\n", err)
	} else {
		conf := new(LoggerConfig)
		err = yaml.Unmarshal(data, conf)
		if err != nil {
			log.Printf("加载日志配置文件失败:%s\n", err)
		}
		return conf
	}
	return nil
}

//添加日志配置
func AddAppender(appender LoggerAppender) {
	switch appender.Adapter {
	case Adapter_File:
		addFileAppender(appender)
		break
	case Adapter_Console:
		addConsoleAppender(appender)
		break
	default:
		fmt.Printf("不能识别的日志适配器:%s", appender.Adapter)
	}
}

//添加console的日志配置
func addConsoleAppender(appender LoggerAppender) {
	addStdOutFilter(getLevel(appender.Level), appender.Package_path, appender.Timeformat, appender.Format)
}

//添加文件系统的日志配置
func addFileAppender(appender LoggerAppender) {
	rotatePolicy := strings.ToLower(appender.Rotate_policy)
	switch rotatePolicy {
	case "time":
		addFileFilterForTime(getLevel(appender.Level), appender.Package_path, appender.Log_path, time.Minute*time.Duration(appender.Rotate_scope), appender.Rotate, appender.Timeformat, appender.Format)
		return
	case "size":
		addFileFilterForSize(getLevel(appender.Level), appender.Package_path, appender.Log_path, appender.Rotate_scope*1048576, appender.Rotate, appender.Timeformat, appender.Format)
		return
	default:
		addFileFilterForDefualt(getLevel(appender.Level), appender.Package_path, appender.Log_path, appender.Timeformat, appender.Format)
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
