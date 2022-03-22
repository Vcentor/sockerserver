// Copyright(C) 2021 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2021/6/7

// https://github.com/go-sql-driver/mysql/blob/master/conncheck.go

// +build linux darwin dragonfly freebsd netbsd openbsd solaris illumos

package pool

import (
	"errors"
	"io"
	"net"
	"syscall"
)

var errUnexpectedRead = errors.New("unexpected read from socket")

// connCheck 检查连接是否有效，若已经无效，会返回 error
//
func connCheck(conn net.Conn) error {
	sysConn, ok := conn.(syscall.Conn)
	if !ok {
		return nil
	}
	rawConn, errSC := sysConn.SyscallConn()
	if errSC != nil {
		return errSC
	}

	var sysErr error
	errRead := rawConn.Read(func(fd uintptr) bool {
		var buf [1]byte
		n, err := syscall.Read(int(fd), buf[:])
		switch {
		case n == 0 && err == nil:
			sysErr = io.EOF
		case n > 0:
			sysErr = errUnexpectedRead
		case err == syscall.EAGAIN || err == syscall.EWOULDBLOCK:
			sysErr = nil
		default:
			sysErr = err
		}
		return true
	})

	if errRead != nil {
		return errRead
	}

	return sysErr
}
