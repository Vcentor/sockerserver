// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Chen Xin (chenxin@baidu.com)
// Date: 2020/04/19

package logit

import (
	"fmt"
	"strings"
)

// Level 日志等级
type Level uint8

// String 日志等级名称， 如 NOTICE
func (l Level) String() string {

	// 这样比直接使用LevelString[l] 性能更好
	switch l {
	case UnknownLevel:
		return unknownLevelName
	case DebugLevel:
		return debugLevelName
	case TraceLevel:
		return traceLevelName
	case NoticeLevel:
		return noticeLevelName
	case WarningLevel:
		return warningLevelName
	case ErrorLevel:
		return errorLevelName
	case FatalLevel:
		return fatalLevelName
	}

	// 兜底，如此也可以自定义扩展新的类型
	return LevelString[l]
}

// Is 是否指定的等级
func (l Level) Is(level Level) bool {
	return l&level == level
}

var (
	// UnknownLevel 未知级别
	UnknownLevel Level // = 0

	// DebugLevel 用于调试的变量的信息数据
	DebugLevel Level = 1 << 0

	// TraceLevel 程序的分支跳转、逻辑流向之类的追踪信息
	TraceLevel Level = 1 << 1

	// NoticeLevel 程序的主要信息，业界一般叫 InfoLevel，百度由于历史原因一直叫 Notice
	NoticeLevel Level = 1 << 2

	// WarningLevel 可能有问题，需要关注的信息
	WarningLevel Level = 1 << 3

	// ErrorLevel 确定出错了的信息
	ErrorLevel Level = 1 << 4

	// FatalLevel 一般记录确定出错的信息
	// ErrorLevel 与 FatalLevel 的区别是 ErrorLevel 表示会话出错，但是程序仍然可以继续运行；
	// FatalLevel 则是程序出错，不能再继续运行了。
	FatalLevel Level = 1 << 5

	// AllLevels 包含所有级别
	AllLevels Level = 1<<8 - 1
)

var (
	// 名称，字符串形式的外部不应该使用，故不输出
	unknownLevelName = "UNKNOWN"
	debugLevelName   = "DEBUG"
	traceLevelName   = "TRACE"
	noticeLevelName  = "NOTICE"
	warningLevelName = "WARNING"
	errorLevelName   = "ERROR"
	fatalLevelName   = "FATAL"
	allLevelName     = "ALL"

	// LevelString 每个 level 的字符串描述
	LevelString = map[Level]string{
		UnknownLevel: unknownLevelName,
		DebugLevel:   debugLevelName,
		TraceLevel:   traceLevelName,
		NoticeLevel:  noticeLevelName,
		WarningLevel: warningLevelName,
		ErrorLevel:   errorLevelName,
		FatalLevel:   fatalLevelName,
		AllLevels:    allLevelName,
	}
)

// ParseLevel 解析字符串的level
func ParseLevel(level string) (Level, error) {
	upper := strings.ToUpper(level)
	for id, name := range LevelString {
		if name == upper {
			return id, nil
		}
	}
	return UnknownLevel, fmt.Errorf("unknown level name %q", level)
}
