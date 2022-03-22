// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2020/7/27

// Package logit 通用日志库
//
// 1.建议将所有日志的配置文件全部放到：conf/logit/ 目录下
//
// 2.配置文件示例（conf/logit/ral.toml）：
//
// 	FileName="{gdp.LogDir}/ral/ral.log"
//
// 	# 日志切分规则,可选参数，默认为1hour
// 	RotateRule="1hour"
//
// 	# 日志文件保留个数，可选参数
// 	# 默认48个，若为-1，日志文件将不清理
// 	MaxFileNum=48
//
// 	# 日志异步队列大小，可选参数
// 	# 默认值 4096，若为-1，则队列大小为0
// 	BufferSize=4096
//
// 	# 日志进入待写队列超时时间，毫秒
// 	# 默认为0，不超时，若出现落盘慢的时候，调用写日志的地方会出现同步等待
// 	#WriterTimeout=0
//
// 	# 日志落盘刷新间隔，毫秒,可选参数
// 	# 若<=0，使用默认值1000
// 	FlushDuration = 2000
//
// 	# 日志编码的对象池名称，可选参数
// 	# 默认为 default_text（普通文本编码）
// 	# 可选值：default_json
// 	# 可通过 RegisterEncoderPool 自定义
// 	EncoderPool="default_text"
//
// 	# 日志内容前缀，可选参数
// 	# 默认为default (包含日志等级、当前时间[精确到秒]、调用位置)
// 	# 可选值：default-默认，时间精确到秒，default_nano-时间精确到纳秒、no-无前缀。
// 	# 可通过 RegisterPrefixFunc 自定义
// 	# Prefix="default"
//
//
// 	# 在logger 的 Output 执行前执行，可选参数
// 	# 可以和 Prefix 配合使用
// 	# 可选值：default-什么都不做，to_body-将level等字段写入日志body
// 	# 可通过 RegisterBeforeOutputFunc 自定义
// 	# BeforeOutput=""
//
// 	[[Dispatch]]
// 	FileSuffix=""
// 	Levels=["NOTICE","TRACE"]
//
// 	[[Dispatch]]
// 	FileSuffix=".wf"
// 	Levels=["WARNING","ERROR","FATAL"]
//
// 3.配置文件示例2(conf/logit/service.toml)：
//
// 	FileName="{gdp.LogDir}/service/service.log"
// 	#其他内容全部使用默认值
//
// 4.使用Logger打印日志
//
// a:一般在app中使用一个独立的文件定义全局的资源对象，如：resource/resource.go
// 	var RalLogger logit.Logger
// b:在APP初始化的时候，初始化完成后将logger赋值给上述全局变量：
// 	func InitLoggers(){
// 		myLogger:=NewSimple(os.Stderr)
// 		resource.RalLogger=myLogger
// 	}
// c:使用的时候：
// 	func myHandler(ctx context.Context){
// 		// 业务逻辑
// 		resource.RalLogger.Notice(ctx,"ok")
// 	}
// 	// 更详细的使用例子请看其他Examples
//
//
// 5.程序在退出前将所有日志落盘：
// 		监听进程退出信号
// 		XXXLogger.(io.Closer).Close() 或者  TryCloseLogger( XXXLogger )
//
// 设计文档：http://wiki.baidu.com/pages/viewpage.action?pageId=1167546359 (使用新页面打开)
package logit
