#logger
======

###获取方式
```
go get github.com/coffeehc/logger
```

###使用方式
coffeehc/logger 是一个基础的日志框架,提供扩展的开放logFilter接口,使用者可以自己定义何种级别的logger发布到对应的io.Writer中

日志级别定义了5个:

1.	trace
2.	debug
3.	info
4.	warn
5.	error


使用配置的方式(yaml语法),配置文件内容如下:
```
context: Default
appenders:
 -
  level: debug
  package_path: /
  adapter: console
  #使用golang自己的timeFormat
  timeformat: 2006-01-02 15:04:05
  format: %T %L %C %M
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
```
系统默认会取-loggerConf参数的值来加载配置文件,如果没有指定则使用debug对所有的包路径下的日志打印到控制台

2015-4-29
1. AddAppender用于自己使用编程方式来定义日志,其实也可以用底层的Filter接口来扩展会更灵活

2015-5-15
在配置中加入了format的参数设置,提供四种标记来组合日志,标记说明如下:

> 1. %T:时间标记,会与timeformat配合使用
> 2. %L:日志级别,这会输出相应的日志级别
> 3. %C:代码信息,这包括包文件描述和日志在第几行打印
> 4. %M:这个就是需要打印的具体日志内容 

###TODO
1. 需要支持没有指定配置文件路径则在程序目录下寻找conf/log.yaml文件来加载的方式,简化启动参数
2. 暂不支持TCP方式存储日志,以后看情况再提供,只要实现io.Writer的接口就可以了,自己动手,丰衣足食
