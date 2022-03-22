// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2020/7/29

package writer

import (
	"io"
)

// Nop 一个黑洞writer
type Nop struct {
}

// Write 写入
// 返回结果总是成功
func (n2 Nop) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// Close 关闭，总是返回nil
func (n2 Nop) Close() error {
	return nil
}

var _ io.WriteCloser = (*Nop)(nil)
