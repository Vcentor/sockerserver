// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2020/8/18

// Package conf 提供配置文件的统一读取能力
//
// 1.默认已经支持 .toml 和 .json 文件的配置文件读取能力，
// 可以通过 RegisterParserFunc 方法注册新的文件类型的支持
//
// 2.扩展能力
//
// 2.1 系统环境变量替换：
// 	# 通过特殊的 {env.xxx} 的语法，可以将系统的环境变量信息替换到配置内容中。
// 	# "|"后面是默认值，可选的
//
// 	http_listen = "0.0.0.0:{env.LISTEN_PORT|8080}"
// 	name = "{env.Name}"
//
// 2.2 GDP应用环境信息变量替换：
// 	# 通过特殊的{gdp.xxx} 语法，可以将gdp/env包设置的环境信息读取到。
// 	log="{gdp.LogDir}/mysql.log"
//
// 	#完整的能力包括 {gdp.RootDir}、{gdp.ConfDir}、{gdp.LogDir}、{gdp.DataDir}、{gdp.IDC}、{gdp.AppName}
//
// 模块设计文档：http://wiki.baidu.com/pages/viewpage.action?pageId=1157157322
package conf
