// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2020/8/21

package writer

import (
	"io"
)

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error {
	return nil
}

// NopCloser  returns a io.WriteCloser with a no-op Close method wrapping
func NopCloser(w io.Writer) io.WriteCloser {
	return nopCloser{w}
}
