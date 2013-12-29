logger
======

coffeehc/logger 是一个基础的日志框架,提供扩展的开放logFilter接口,使用者可以自己定义何种级别的logger发布到对应的io.Writer中

日志级别定义了5个:

1.	trace
2.	debug
3.	info
4.	warn
5.	error


提供下面几种日志方式
```
	func AddStdOutFilter(name string, level byte, path string, timeFormat string)
	AddFileFilterForTime("testTimeLog", LOGGER_LEVEL_DEBUG, "/", "d:/testlog/testLog.log", time.Second*10, 10)
	AddFileFilterForDefualt("testTimeLog", LOGGER_LEVEL_DEBUG, "/", "d:/testlog/testLog.log")
	AddFileFilterForSize("testTimeLog", LOGGER_LEVEL_DEBUG, "/", "d:/testlog/testLog.log", 3*1024, 10)
```

暂不支持TCP方式存储日志,以后看情况再提供,只要实现io.Writer的接口就可以了,自己动手,丰衣足食