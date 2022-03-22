// Copyright(C) 2021 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2021/10/14

package writer

import (
	"io"
)

// NewLineBuffer 创建一个 buffer size =4096 的 LineBuffer
func NewLineBuffer(w io.Writer) *LineBuffer {
	return NewLineBufferSize(w, 4096)
}

// NewLineBufferSize 创建一个 指定 buffer size 的 LineBuffer
func NewLineBufferSize(w io.Writer, size int) *LineBuffer {
	if size <= 0 {
		size = 4096
	}
	return &LineBuffer{
		wr:  w,
		buf: make([]byte, size),
	}
}

// LineBuffer 按照'行'缓存的 buffer writer
// 这个'行' 指的是 一次调用 Write() 的内容，
// 一般在写日志的时候，都是将一整行(已\n结尾的字符串)内容发送给 Writer
//
// 复用了 bufio.Writer 的代码,和 bufio.Writer 的区别是：
// 	bufio.Writer 是严格按照 buffer size 来缓存的，默认是 4096，
// 	也就是当缓存的数据达到 4096 后，会将数据 Flush,同时 buffer 里会有剩余部分，如
// 	假设 buffer size=4，buf="ab", b.Write([]byte("cde"))
// 	会将内容 "abcd" Flush 出去，buf="e"
//
// 	而当前 LineBuffer 的实现则是：
// 	若新写入的和已 buffered 内容超过 buffer size，会先 Flush，
// 	同时，若剩余 buffer 长度小于当前写入内容大小，也会将当前内容 Flush
// 	假设 buffer size=4，buf="ab", b.Write([]byte("cde"))
// 	会先将 "ab" Flush 出去，然后将 "cde" 也 Flush 出去
type LineBuffer struct {
	wr io.Writer

	buf []byte
	n   int
	err error
}

// Write 写入
func (lb *LineBuffer) Write(p []byte) (nn int, err error) {
	if lb.err != nil {
		return 0, lb.err
	}
	for len(p) > lb.Available() && lb.err == nil {
		if lb.Buffered() == 0 {
			var n int
			n, lb.err = lb.wr.Write(p)
			nn += n
			p = p[n:]
		} else {
			lb.Flush()
		}

		if lb.err != nil {
			return nn, lb.err
		}
	}
	n := copy(lb.buf[lb.n:], p)
	lb.n += n
	nn += n

	// buffer 快满了
	if len(p) > lb.Available() {
		lb.Flush()
	}
	return nn, lb.err
}

// Reset 重置 writer，同时也会将 err,buf 重置
func (lb *LineBuffer) Reset(w io.Writer) {
	lb.err = nil
	lb.n = 0
	lb.wr = w
}

// Available 剩余可用 buffer 大小
func (lb *LineBuffer) Available() int {
	return len(lb.buf) - lb.n
}

// Buffered 已缓存的大小
func (lb *LineBuffer) Buffered() int {
	return lb.n
}

// Flush 将内容写入的 实际的 writer
func (lb *LineBuffer) Flush() error {
	if lb.err != nil {
		return lb.err
	}
	if lb.n == 0 {
		return nil
	}
	n, err := lb.wr.Write(lb.buf[0:lb.n])
	if n < lb.n && err == nil {
		err = io.ErrShortWrite
	}
	if err != nil {
		if n > 0 && n < lb.n {
			copy(lb.buf[0:lb.n-n], lb.buf[n:lb.n])
		}
		lb.n -= n
		lb.err = err
		return err
	}
	lb.n = 0
	return nil
}
