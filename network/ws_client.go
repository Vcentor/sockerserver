// @Author: Vcentor
// @Date: 2020/12/3 4:36 下午

package network

import (
	"context"
	"socketserver/library/wslog"
	"errors"
	"github.com/gorilla/websocket"
	"icode.baidu.com/baidu/gdp/logit"
	"sync"
	"time"
)

const (
	RECONNECT_SLEEP_DURATION = 50
)

// WSClient websocket client
type WSClient struct {
	addr      string
	chanCap   int
	conn      *websocket.Conn
	readChan  chan readChan
	writeChan chan WriteChan
	closeFlag bool
	closeChan chan byte
	mutex     sync.Mutex
}

// NewWSClient 初始化websocket client
func NewWSClient(addr string, chanCap, retry int) *WSClient {
	var (
		conn *websocket.Conn
		err  error
	)
	for i := 0; i < retry; i++ {
		conn, _, err = websocket.DefaultDialer.Dial(addr, nil)
		if err != nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		break
	}
	if conn == nil {
		wslog.Logger.Fatal(context.Background(), "WSClient connect fail", logit.String("addr", addr), logit.Error("error", err))
		return nil
	}

	wsClient := &WSClient{
		addr:      addr,
		chanCap:   chanCap,
		conn:      conn,
		readChan:  make(chan readChan, chanCap),
		writeChan: make(chan WriteChan, chanCap),
		closeFlag: false,
		closeChan: make(chan byte, 1),
	}
	go wsClient.readLoop()
	go wsClient.writeLoop()
	return wsClient
}

// ReadMsg 读取数据
func (wc *WSClient) ReadMsg() (data []byte, messageType int, err error) {
	select {
	case rc := <-wc.readChan:
		data = rc.message
		messageType = rc.messageType
	case <-wc.closeChan:
		err = errors.New("WSClient channel is closed")
	}
	return
}

// readLoop	读取数据
func (wc *WSClient) readLoop() {
	for {
		messageType, data, err := wc.conn.ReadMessage()
		if err != nil {
			wslog.Logger.Notice(context.Background(), "WSClient read message failed", logit.Error("error", err))
			goto CLOSE
		}
		var rc = readChan{
			message:     data,
			messageType: messageType,
		}
		select {
		case wc.readChan <- rc:
		case <-wc.closeChan:
			goto CLOSE
		}
	}
CLOSE:
	wc.Close()
}

// WriteMsg 发送websocket信息
func (wc *WSClient) WriteMsg(data WriteChan) (err error) {
	select {
	case wc.writeChan <- data:
	case <-wc.closeChan:
		err = errors.New("WSClient channel is closed")
	}
	return
}

// writeLoop 发送websocket信息
func (wc *WSClient) writeLoop() {
	var data WriteChan
	for {
		select {
		case data = <-wc.writeChan:
			if err := wc.conn.WriteMessage(data.Type, data.Message); err != nil {
				goto CLOSE
			}

		case <-wc.closeChan:
			goto CLOSE
		}
	}
CLOSE:
	wc.Close()
}

// Close 关闭连接
func (wc *WSClient) Close() {
	if !wc.closeFlag {
		wc.mutex.Lock()
		if !wc.closeFlag {
			wc.conn.Close()
			close(wc.closeChan)
			wc.closeFlag = true
		}
		wc.mutex.Unlock()
	}
}

// ReConnect 断线重连
func (wc *WSClient) ReConnect() {
	wc.mutex.Lock()
	if wc.closeFlag {
	reconnect:
		conn, _, err := websocket.DefaultDialer.Dial(wc.addr, nil)
		if err != nil {
			wslog.Logger.Fatal(context.Background(), "WSClient reconnect fail", logit.String("addr", wc.addr), logit.Error("error", err))
			time.Sleep(RECONNECT_SLEEP_DURATION * time.Millisecond)
			goto reconnect
		}
		wc.conn = conn
		wc.closeFlag = false
		wc.closeChan = make(chan byte, 1)
		wc.readChan = make(chan readChan, wc.chanCap)
		wc.writeChan = make(chan WriteChan, wc.chanCap)
		go wc.readLoop()
		go wc.writeLoop()
		wslog.Logger.Debug(context.Background(), "reconnect success")
	}
	wc.mutex.Unlock()
}

// GetCloseFlag 获取closeFlag
func (wc *WSClient) GetCloseFlag() bool {
	wc.mutex.Lock()
	flag := wc.closeFlag
	wc.mutex.Unlock()
	return flag
}
