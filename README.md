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

```
系统默认会取-loggerConf参数的值来加载配置文件,如果没有指定则使用debug对所有的包路径下的日志打印到控制台

###TODO
1. 需要支持没有指定配置文件路径则在程序目录下寻找log.yaml文件来加载的方式,简化启动参数
2. 暂不支持TCP方式存储日志,以后看情况再提供,只要实现io.Writer的接口就可以了,自己动手,丰衣足食