// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2020/8/19

package b2log

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
	"unsafe"

	// nolint
	"github.com/golang/protobuf/proto"
)

// magic number
// This can only work in little-endian machine (e.g., x86)
// Remember to make MAGIC_NUMBER and MAGIC_NUMBER_STR consistent
const (
	MagicNumber   = 0xB0AEBEA7
	HeaderVersion = 1
)

// nolint
var demoHeader Header // this var is only for getting size of Header

// HeaderSize 日志头长度
var HeaderSize = int(unsafe.Sizeof(demoHeader))

// Header for b2log record
type Header struct {
	MagicNumber   uint32 // magic number
	Version       uint32 // version
	UnCompressLen uint32 // length of unCompress log
	CompressLen   uint32 // length of compress log
	TimeStamp     uint64 // timestamp the log generated
}

// NowFunc 获取当前时间
var NowFunc = time.Now

// generate timestamp
func timestampGen() uint64 {
	t := NowFunc()
	sec := t.Unix()
	usec := t.Nanosecond() / 1000
	ts := uint64(sec*1000 + int64(usec)/1000)
	return ts
}

// Encode 将一条pb日志信息打包为bytes
func Encode(msg proto.Message) ([]byte, error) {
	body, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	hd, err := genHeader(body)
	if err != nil {
		return nil, err
	}

	bf := make([]byte, len(hd)+len(body))
	copy(bf, hd)
	copy(bf[len(hd):], body)
	return bf, nil
}

func genHeader(body []byte) ([]byte, error) {
	header := Header{
		MagicNumber:   MagicNumber,
		Version:       HeaderVersion,
		UnCompressLen: uint32(len(body)),
		CompressLen:   0,
		TimeStamp:     timestampGen(),
	}

	bf := new(bytes.Buffer)
	err := binary.Write(bf, binary.LittleEndian, header)
	if err != nil {
		return nil, fmt.Errorf("binary.Write() failed: %w", err)
	}

	return bf.Bytes(), nil
}
