// @Author: Vcentor
// @Date: 2020/9/24 10:58 上午

package network

import (
	"context"
	"socketserver/library/wslog"
	"errors"
	"github.com/gorilla/websocket"
	"icode.baidu.com/baidu/gdp/logit"
	"net"
	"sync"
	"time"
)

// 心跳超时，15s不发心跳就自动断开
const HEARTBEAT_TIMEOUT = 15

// WSConnSet 连接池
type WSConnPool map[*WSConn]string

// readChan 读channel
type readChan struct {
	message     []byte
	messageType int
}

// WriteChan 写channel
type WriteChan struct {
	Message []byte
	Type    int
}

// WSConn websocket connection
// read message and write message
type WSConn struct {
	conn      *websocket.Conn
	writeChan chan []byte
	readChan  chan readChan
	closeChan chan byte
	closeFlag bool
	mutex     sync.Mutex
	handler   *WSHandler
	sessionID string
	authInfo  map[string]bool
}

// newWSConn 初始化WSConn
func newWSConn(conn *websocket.Conn, handler *WSHandler, chanCap int, ssid string) *WSConn {
	var wsConn = &WSConn{
		conn:      conn,
		writeChan: make(chan []byte, chanCap),
		readChan:  make(chan readChan, chanCap),
		closeChan: make(chan byte, 1),
		closeFlag: false,
		handler:   handler,
		sessionID: ssid,
		authInfo:  make(map[string]bool),
	}

	go wsConn.readLoop()
	go wsConn.writeLoop()
	return wsConn
}

func (wsConn *WSConn) ReadMsg() (data []byte, messageType int, err error) {
	select {
	case rc := <-wsConn.readChan:
		data = rc.message
		messageType = rc.messageType
	case <-time.After(HEARTBEAT_TIMEOUT * time.Second):
		wsConn.WriteMsg([]byte(`read timeout connect closed`))
		err = errors.New("heartbeat timeout, connection is closed")
	case <-wsConn.closeChan:
		err = errors.New("connection is closed")
	}
	return
}

// readLoop channel接收数据，协程安全
func (wsConn *WSConn) readLoop() {
	for {
		messageType, data, err := wsConn.conn.ReadMessage()
		if err != nil {
			wslog.Logger.Notice(context.Background(), "WSConn read message failed", logit.Error("error", err))
			goto CLOSE
		}
		var rc = readChan{
			message:     data,
			messageType: messageType,
		}
		select {
		case wsConn.readChan <- rc:
		case <-wsConn.closeChan:
			goto CLOSE
		}
	}
CLOSE:
	wsConn.Close()
}

// WriteMsg 发送数据
func (wsConn *WSConn) WriteMsg(b []byte) (err error) {
	select {
	case wsConn.writeChan <- b:
	case <-wsConn.closeChan:
		err = errors.New("channel is closed")
	}
	return
}

func (wsConn *WSConn) writeLoop() {
	var data []byte
	for {
		select {
		case data = <-wsConn.writeChan:
		case <-wsConn.closeChan:
			goto CLOSE
		}
		if err := wsConn.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			wslog.Logger.Notice(context.Background(), "WSConn write text message failed", logit.String("data", string(data)), logit.Error("error", err))
			goto CLOSE
		}
	}
CLOSE:
	wsConn.Close()
}

func (wsConn *WSConn) Close() {
	if !wsConn.closeFlag {
		// 线程安全的，可重复调用
		wsConn.conn.Close()
		// channel只能关闭一次，且非线程安全
		wsConn.handler.mutex.Lock()
		if !wsConn.closeFlag {
			// 删除鉴权信息,释放内存
			wsConn.DelAuthInfo(wsConn.handler.conns[wsConn])
			// 删除连接池，释放内存
			delete(wsConn.handler.conns, wsConn)
			close(wsConn.closeChan)
			wsConn.closeFlag = true
		}
		wsConn.handler.mutex.Unlock()
	}
}

// LocalAddr 本机地址
func (wsConn *WSConn) LocalAddr() net.Addr {
	return wsConn.conn.LocalAddr()
}

// RemoteAddr 远程地址
func (wsConn *WSConn) RemoteAddr() net.Addr {
	return wsConn.conn.RemoteAddr()
}

// GetConnPool 获取连接池
func (wsConn *WSConn) GetConnPool() []*WSConn {
	var conns []*WSConn
	wsConn.handler.mutex.Lock()
	for conn := range wsConn.handler.conns {
		conns = append(conns, conn)
	}
	wsConn.handler.mutex.Unlock()
	return conns
}

// GetSessionPool 获取所有连接池中的session
func (wsConn *WSConn) GetSessionPool() []string {
	var sessionIDs []string
	wsConn.handler.mutex.Lock()
	for _, sid := range wsConn.handler.conns {
		sessionIDs = append(sessionIDs, sid)
	}
	return sessionIDs
}

// GetCloseFlag 获取关闭标记
func (wsConn *WSConn) GetCloseFlag() bool {
	return wsConn.closeFlag
}

// GetSessionID 获取session id
func (wsConn *WSConn) GetSessionID() string {
	return wsConn.sessionID
}

// SetAuthInfo 设置鉴权信息
func (wsConn *WSConn) SetAuthInfo(sid string) {
	wsConn.mutex.Lock()
	wsConn.authInfo[sid] = true
	wsConn.mutex.Unlock()
}

// 删除鉴权信息
func (wsConn *WSConn) DelAuthInfo(sid string) {
	wsConn.mutex.Lock()
	delete(wsConn.authInfo, sid)
	wsConn.mutex.Unlock()
}

// Auth 鉴权
func (wsConn *WSConn) Auth(sid string) bool {
	var res bool
	wsConn.mutex.Lock()
	if _, ok := wsConn.authInfo[sid]; ok {
		res = true
	}
	wsConn.mutex.Unlock()
	return res
}
