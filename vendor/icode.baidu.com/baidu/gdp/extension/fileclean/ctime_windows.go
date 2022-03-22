// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2020/7/24

// +build windows

package fileclean

import (
	"os"
)

func ctime(info os.FileInfo) int64 {
	return info.ModTime().UnixNano()
}
