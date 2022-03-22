// Author: Vcentor
// Date: 2022/3/17 7:09 下午
// desc:

package network

import (
	"context"
	"socketserver/library/wslog"
	"icode.baidu.com/baidu/gdp/logit"
	"net"
	"sync"
)

// TCPConnPool 连接池
type TCPConnPool struct {
	Pool map[*TCPConn]string
	sync.RWMutex
}

// NewTCPConnPool 实例化对象
func NewTCPConnPool() *TCPConnPool {
	return &TCPConnPool{
		Pool: make(map[*TCPConn]string),
	}
}

//ConnNum 连接数
func (pool *TCPConnPool) Len() int {
	pool.Lock()
	defer pool.Unlock()
	return len(pool.Pool)
}

// WithConn 加入连接
func (pool *TCPConnPool) WithConn(conn *TCPConn, ssid string) {
	pool.Lock()
	defer pool.Unlock()
	pool.Pool[conn] = ssid
}

// GetConnBySsid 通过ssid获取连接
func (pool *TCPConnPool) GetConnBySsid(ssid string) *TCPConn {
	pool.RLock()
	defer pool.RUnlock()
	for conn, id := range pool.Pool {
		if id == ssid {
			return conn
		}
	}
	return nil
}

// SessionID
func (pool *TCPConnPool) SessionID(conn *TCPConn) string {
	pool.RLock()
	defer pool.RUnlock()
	sid, ok := pool.Pool[conn]
	if !ok {
		return ""
	}
	return sid
}

// GetConns 获取所有连接
func (pool *TCPConnPool) GetConns() []*TCPConn {
	pool.RLock()
	defer pool.RUnlock()
	var conns = make([]*TCPConn, 0)
	for conn := range pool.Pool {
		conns = append(conns, conn)
	}
	return conns
}

// GetSessionIDs 获取所有sessionID
func (pool *TCPConnPool) GetSessionIDs() []string {
	pool.RLock()
	defer pool.RUnlock()
	var ssids = make([]string, 0)
	for _, id := range pool.Pool {
		ssids = append(ssids, id)
	}
	return ssids
}

// DelConn 回收连接释放内存
func (pool *TCPConnPool) DelConn(conn *TCPConn) {
	pool.Lock()
	defer pool.Unlock()
	delete(pool.Pool, conn)
}

// TCPConn read and write
type TCPConn struct {
	sync.Mutex
	conn      net.Conn
	writeChan chan []byte
	connPool  *TCPConnPool
	closeFlag bool
	closeChan chan byte
	parser    *TCPParser
	sessionID string
}

// newTCPConn 初始化TCPConn
func newTCPConn(conn net.Conn, chanCap int, ssid string, parser *TCPParser, pool *TCPConnPool) *TCPConn {
	tcpConn := &TCPConn{
		conn:      conn,
		writeChan: make(chan []byte, chanCap),
		connPool:  pool,
		closeFlag: false,
		closeChan: make(chan byte, 1),
		parser:    parser,
		sessionID: ssid,
	}
	go tcpConn.writeLoop()
	return tcpConn
}

// Read data
func (tcpConn *TCPConn) Read(b []byte) (n int, err error) {
	return tcpConn.conn.Read(b)
}

func (tcpConn *TCPConn) doWrite(b []byte) {
	if len(tcpConn.writeChan) == cap(tcpConn.writeChan) {
		wslog.Logger.Notice(context.Background(), "close tcp conn: channel full")
		tcpConn.Close()
	}
	tcpConn.writeChan <- b
}

// Write data
func (tcpConn *TCPConn) Write(b []byte) {
	tcpConn.Lock()
	defer tcpConn.Unlock()
	if tcpConn.closeFlag || b == nil {
		return
	}
	tcpConn.doWrite(b)
}

// writeLoop
func (tcpConn *TCPConn) writeLoop() {
	var b []byte
	for {
		select {
		case b = <-tcpConn.writeChan:
		case <-tcpConn.closeChan:
			goto CLOSE
		}
		if _, err := tcpConn.conn.Write(b); err != nil {
			wslog.Logger.Notice(context.Background(), "TCPConn write message failed!", logit.String("message", string(b)), logit.Error("error", err))
			goto CLOSE
		}
	}
CLOSE:
	tcpConn.Close()
}

// Close 关闭连接，释放内存
func (tcpConn *TCPConn) Close() {
	if !tcpConn.closeFlag {
		tcpConn.Lock()
		if !tcpConn.closeFlag {
			_ = tcpConn.conn.(*net.TCPConn).SetLinger(0)
			_ = tcpConn.conn.Close()
			tcpConn.connPool.DelConn(tcpConn)
			close(tcpConn.closeChan)
			tcpConn.closeFlag = true
		}
		tcpConn.Unlock()
	}
}

// LocalAddr 本机地址
func (tcpConn *TCPConn) LocalAddr() net.Addr {
	return tcpConn.conn.LocalAddr()
}

// RemoteAddr 远程地址
func (tcpConn *TCPConn) RemoteAddr() net.Addr {
	return tcpConn.conn.RemoteAddr()
}

// GetConnPool 获取所有连接
func (tcpConn *TCPConn) GetConnPool() []*TCPConn {
	return tcpConn.connPool.GetConns()
}

// GetSessionPool 获取sessionIDs
func (tcpConn *TCPConn) GetSessionPool() []string {
	return tcpConn.connPool.GetSessionIDs()
}

// GetCloseFlag 获取关闭channel标识
func (tcpConn *TCPConn) GetCloseFlag() bool {
	return tcpConn.closeFlag
}

// GetSessionID 获取session信息
func (tcpConn *TCPConn) GetSessionID() string {
	return tcpConn.connPool.SessionID(tcpConn)
}

// ReadMsg 根据协议读取数据
func (tcpConn *TCPConn) ReadMsg() ([]byte, error) {
	return tcpConn.parser.Read(tcpConn)
}

// WriteMsg 根据协议写入数据
func (tcpConn *TCPConn) WriteMsg(args ...[]byte) error {
	return tcpConn.parser.Write(tcpConn, args...)
}
