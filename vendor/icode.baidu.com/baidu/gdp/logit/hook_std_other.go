// Copyright(C) 2021 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2021/12/6

//go:build !windows && !darwin
// +build !windows,!darwin

package logit

import (
	"syscall"
)

// HookStderr 劫持标准错误输出
// 	当前程序的 stderr 的内容将输出指定的文件
func HookStderr(f HasFd) error {
	return syscall.Dup3(int(f.Fd()), 2, syscall.O_CLOEXEC)
}

// HookStdout 劫持标准输出
// 	当前程序的 stdout 的内容将输出指定的文件
func HookStdout(f HasFd) error {
	return syscall.Dup3(int(f.Fd()), 1, syscall.O_CLOEXEC)
}
