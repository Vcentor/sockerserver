// Author: Vcentor
// Date: 2022/3/17 7:32 下午
// desc:

package network

import (
	"bufio"
	"encoding/binary"
	"errors"
	"math"
)

// TCPParser  encode or decode data
// len | data
type TCPParser struct {
	lenMsgLen    int    // message长度所占的字节数,默认2个字节
	minMsgLen    uint32 // 最小message长度,默认1个字节
	maxMsgLen    uint32 // 最大message长度,默认4096个字节
	littleEndian bool   // 字节序,默认大端序
}

// NewTCPParser 实例化TCPParser
func NewTCPParser() *TCPParser {
	return &TCPParser{
		lenMsgLen:    2,
		minMsgLen:    1,
		maxMsgLen:    4096,
		littleEndian: false,
	}
}

// WithMsgLen 自定义数据长度
func (p *TCPParser) WithMsgLen(lenMsgLen int, minMsgLen, maxMsgLen uint32) {
	if lenMsgLen == 1 || lenMsgLen == 2 || lenMsgLen == 4 {
		p.lenMsgLen = lenMsgLen
	}

	if minMsgLen != 0 {
		p.minMsgLen = minMsgLen
	}

	if maxMsgLen != 0 {
		p.maxMsgLen = maxMsgLen
	}

	var max uint32
	switch lenMsgLen {
	case 1:
		max = math.MaxUint8
	case 2:
		max = math.MaxUint16
	case 4:
		max = math.MaxUint32
	}

	if p.minMsgLen > max {
		p.minMsgLen = max
	}

	if p.maxMsgLen > max {
		p.maxMsgLen = max
	}
}

// WithEndian 设置字节序
func (p *TCPParser) WithEndian(littleEndian bool) {
	p.littleEndian = littleEndian
}

// Read 读取信息
func (p *TCPParser) Read(conn *TCPConn) ([]byte, error) {
	reader := bufio.NewReader(conn)
	bufMsgLen, err := reader.Peek(p.lenMsgLen)
	if err != nil {
		return nil, err
	}

	// parse len
	var msgLen uint32
	switch p.lenMsgLen {
	case 1:
		msgLen = uint32(bufMsgLen[0])
	case 2:
		if p.littleEndian {
			msgLen = uint32(binary.LittleEndian.Uint16(bufMsgLen))
		} else {
			msgLen = uint32(binary.BigEndian.Uint16(bufMsgLen))
		}
	case 4:
		if p.littleEndian {
			msgLen = binary.LittleEndian.Uint32(bufMsgLen)
		} else {
			msgLen = binary.BigEndian.Uint32(bufMsgLen)
		}
	}

	// check len
	if msgLen > p.maxMsgLen {
		return nil, errors.New("message too long")
	}

	if msgLen < p.minMsgLen {
		return nil, errors.New("message too short")
	}

	msgData := make([]byte, uint32(p.lenMsgLen)+msgLen)
	if _, err := reader.Read(msgData); err != nil {
		return nil, err
	}

	return msgData[p.lenMsgLen:], nil
}

// Write 写入信息
func (p *TCPParser) Write(conn *TCPConn, args ...[]byte) error {
	var msgLen uint32
	for i := 0; i < len(args); i++ {
		msgLen += uint32(len(args[i]))
	}

	if msgLen > p.maxMsgLen {
		return errors.New("message too long")
	}

	if msgLen < p.minMsgLen {
		return errors.New("message too short")
	}

	var msg = make([]byte, uint32(p.lenMsgLen)+msgLen)
	switch p.lenMsgLen {
	case 1:
		msg[0] = byte(msgLen)
	case 2:
		if p.littleEndian {
			binary.LittleEndian.PutUint16(msg, uint16(msgLen))
		} else {
			binary.BigEndian.PutUint16(msg, uint16(msgLen))
		}
	case 4:
		if p.littleEndian {
			binary.LittleEndian.PutUint32(msg, msgLen)
		} else {
			binary.BigEndian.PutUint32(msg, msgLen)
		}
	}

	var l = p.lenMsgLen
	for i := 0; i < len(args); i++ {
		copy(msg[l:], args[i])
		l += len(args[i])
	}

	conn.Write(msg)

	return nil
}
