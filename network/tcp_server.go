// Author: Vcentor
// Date: 2022/3/17 6:51 下午
// desc:

package network

import (
	"context"
	"socketserver/library/utils"
	"socketserver/library/wslog"
	"fmt"
	"log"
	"net"
	"time"
)

// TCPServer start and run
type TCPServer struct {
	Ctx        context.Context
	ListenAddr string
	MaxConnNum int
	ChanCap    int
	NewAgent   func(*TCPConn) Agent
	ln         net.Listener
	connPool   *TCPConnPool

	// parser
	LenMsgLen    int
	MaxMsgLen    uint32
	MinMsgLen    uint32
	LittleEndian bool
	tcpParser    *TCPParser
}

// Start 启动
func (tcpServer *TCPServer) Start() {
	ctx, _ := context.WithCancel(tcpServer.Ctx)
	tcpServer.init()
	go tcpServer.run(ctx)
}

// init 初始化
func (tcpServer *TCPServer) init() {
	ln, err := net.Listen("tcp", tcpServer.ListenAddr)
	if err != nil {
		panic(err)
	}

	tcpServer.ln = ln

	if tcpServer.MaxConnNum <= 0 {
		tcpServer.MaxConnNum = 100
		log.Printf("Invalid MaxConnNum, reset to %d\n", tcpServer.MaxConnNum)
	}

	if tcpServer.ChanCap <= 0 {
		tcpServer.ChanCap = 100
		log.Printf("Invalid ChanCap, reset to %d\n", tcpServer.ChanCap)
	}

	tcpServer.connPool = NewTCPConnPool()

	tcpParser := NewTCPParser()
	tcpParser.WithMsgLen(tcpServer.LenMsgLen, tcpServer.MinMsgLen, tcpServer.MaxMsgLen)
	tcpParser.WithEndian(tcpParser.littleEndian)

	tcpServer.tcpParser = tcpParser
}

// run 启动
func (tcpServer *TCPServer) run(ctx context.Context) {
	log.Println("TCP Server start...")
	log.Println("TCP address [" + tcpServer.ListenAddr + "]")

	var tempDelay time.Duration
	for {
		conn, err := tcpServer.ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}

				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				wslog.Logger.Notice(ctx, fmt.Sprintf("TCPAccept:accept error: %v; retrying in %v", err, tempDelay))

				time.Sleep(tempDelay)
				continue
			}
			return
		}
		tempDelay = 0

		if tcpServer.connPool.Len() >= tcpServer.MaxConnNum {
			_ = conn.Close()
			wslog.Logger.Notice(ctx, "TCPAccept:too many connects")
			continue
		}

		ssid := utils.NewUUID()
		netConn := newTCPConn(conn, tcpServer.ChanCap, ssid, tcpServer.tcpParser, tcpServer.connPool)
		tcpServer.connPool.WithConn(netConn, ssid)

		agent := tcpServer.NewAgent(netConn)
		// 不能阻塞掉，否则无法接收连接
		go agent.ReadMsg()
	}
}

// Close 关闭
func (tcpServer *TCPServer) Close() {
	_ = tcpServer.ln.Close()
	for _, conn := range tcpServer.connPool.GetConns() {
		conn.Close()
	}
}
