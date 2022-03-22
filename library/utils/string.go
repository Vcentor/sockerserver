// @Author: Vcentor
// @Date: 2020/10/20 5:03 下午

package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// Md5String 字符串md5值
func Md5String(s string) string {
	h := md5.New()
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Md5File 文件md5值
func Md5File(filename string) (string, error) {
	f, err := os.Open(filename)
	defer f.Close()
	if err != nil {
		return "", err
	}
	h := md5.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// NewUUID V4即随机数生成的uuid
func NewUUID() string {
	var buf [16]byte
	for {
		if _, err := rand.Read(buf[:]); err == nil {
			break
		}
	}
	// 设置uuid的版本信息
	buf[6] = (buf[6] & 0x0f) | (4 << 4) // Version 4
	buf[8] = (buf[8] & 0xbf) | 0x80     //  Variant is 10

	res := make([]byte, 36)
	hex.Encode(res[0:8], buf[0:4])
	res[8] = '-'
	hex.Encode(res[9:13], buf[4:6])
	res[13] = '-'
	hex.Encode(res[14:18], buf[6:8])
	res[18] = '-'
	hex.Encode(res[19:23], buf[8:10])
	res[23] = '-'
	hex.Encode(res[24:], buf[10:])
	return string(res)
}
