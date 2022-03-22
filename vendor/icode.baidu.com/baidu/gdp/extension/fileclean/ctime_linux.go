// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2020/7/24

// +build linux

package fileclean

import (
	"os"
	"syscall"
)

func ctime(info os.FileInfo) int64 {
	stat := info.Sys().(*syscall.Stat_t)
	return int64(stat.Ctim.Sec)
}
