// Copyright(C) 2021 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2021/6/7

// https://github.com/go-sql-driver/mysql/blob/master/conncheck_dummy.go

// +build !linux,!darwin,!dragonfly,!freebsd,!netbsd,!openbsd,!solaris,!illumos

package pool

import "net"

func connCheck(conn net.Conn) error {
	return nil
}
