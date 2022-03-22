// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2020/8/18

// Package env 该模块定义APP应用所需要的一些运行环境相关的基础信息。
//
// 如:
// 	应用名称（AppName)     [如：searchbox]
// 	当前机房(IDC)          [如：jx,tc,gz]
// 	运行模式（RunMode）     [可选：debug、test、release(默认)]
// 	应用根目录(RootDir)
// 	配置文件目录(ConfDir)
// 	数据根目录(DataDir)
//
// 应用在启动阶段，首先将这些公共的基础信息进行初始化，其他功能模块则可以直接使用，以达到解耦合的目的。
// 推荐将环境信息配置到 conf/app.toml 文件中：
// 	APPName="demo"
//
// 	IDC="{env.IDC|test}"
//
// 	# RunMode 可选 debug、test、release
// 	RunMode="debug"
//
// 模块设计文档：http://wiki.baidu.com/pages/viewpage.action?pageId=1157831234
package env
