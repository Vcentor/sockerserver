// Copyright(C) 2021 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2021/12/6

package logit

import (
	"io"
	"log"
	"path/filepath"

	"icode.baidu.com/baidu/gdp/env"
)

// HasFd 有实现 Fd 方法
type HasFd interface {
	Fd() uintptr
}

// StdHooker 封装对 Stderr 和 Stdout 内容的劫持输出的能力
type StdHooker struct {
	// StderrConfName stderr 内容输出的配置，可选
	// 为空时会尝试使用 logit/stderr.toml
	StderrConfName string

	// StdoutConfName stdout 内容输出的配置，可选
	// 为空时会尝试使用 logit/stdout.toml
	StdoutConfName string

	stderrWriter io.WriteCloser
	stdoutWriter io.WriteCloser
}

// HookStd 同时劫持 Stderr 和 Stdout
func (hs *StdHooker) HookStd() {
	hs.HookStderr()
	hs.HookStdout()
}

// HookStderr 劫持 Stderr
// 若是配置文件异常，会忽略错误，并使用 log.Println 打印日志
func (hs *StdHooker) HookStderr() {
	confName := hs.StderrConfName
	if len(confName) == 0 {
		confName = "logit/stderr.toml"
	}
	opts := []Option{
		OptConfigFileOrLogFileName(confName, filepath.Join(env.LogDir(), "std", "stderr.log")),
	}
	cfg, err := loggerCfg(opts...)
	if err != nil {
		log.Println("HookStderr loggerCfg failed:", err.Error())
		return
	}
	cfg.addEventListen(func(w io.Writer) {
		if fd, ok := w.(HasFd); ok {
			errHook := HookStderr(fd)
			log.Println("HookStderr", errHook)
		} else {
			log.Println("HookStderr failed, Writer not implement Fd()uintptr")
		}
	})
	w, err := cfg.getWriter()
	if err != nil {
		log.Println("HookStderr getWriter failed:", err.Error())
		return
	}
	hs.stderrWriter = w
}

// HookStdout 劫持 Stdout
// 若是配置文件异常，会忽略错误，并使用 log.Println 打印日志
func (hs *StdHooker) HookStdout() {
	confName := hs.StdoutConfName
	if len(confName) == 0 {
		confName = "logit/stdout.toml"
	}
	opts := []Option{
		OptConfigFileOrLogFileName(confName, filepath.Join(env.LogDir(), "std", "stdout.log")),
	}
	cfg, err := loggerCfg(opts...)
	if err != nil {
		log.Println("HookStdout loggerCfg failed:", err.Error())
		return
	}
	cfg.addEventListen(func(w io.Writer) {
		if fd, ok := w.(HasFd); ok {
			errHook := HookStdout(fd)
			log.Println("HookStdout", errHook)
		} else {
			log.Println("HookStdout failed, Writer not implement Fd()uintptr")
		}
	})
	w, err := cfg.getWriter()
	if err != nil {
		log.Println("HookStdout getWriter failed:", err.Error())
		return
	}
	hs.stdoutWriter = w
}

// Close 关闭底层的 writer
func (hs *StdHooker) Close() {
	if hs.stderrWriter != nil {
		_ = hs.stderrWriter.Close()
	}
	if hs.stdoutWriter != nil {
		_ = hs.stdoutWriter.Close()
	}
}
