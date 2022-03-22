// @Author: Vcentor
// @Date: 2020/9/24 10:58 上午

package network

import (
	"context"
	"crypto/tls"
	"socketserver/library/utils"
	"socketserver/library/wslog"
	"icode.baidu.com/baidu/gdp/logit"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSServer websocket server run
type WSServer struct {
	Ctx         context.Context
	HTTPServer  *http.Server
	Addr        string
	MaxConnNum  int
	WriteMsgCap int
	HTTPTimeout time.Duration
	CerFile     string
	KeyFile     string
	NewAgent    func(*WSConn) Agent
	FailChan    chan error
	handler     *WSHandler
	ln          net.Listener
	// ReadMaxMsgLen uint32
}

// WSHandler handle tcp to websocket
type WSHandler struct {
	ctx         context.Context
	maxConnNum  int
	writeMsgCap int
	upgrader    websocket.Upgrader
	conns       WSConnPool
	mutex       sync.Mutex
	newAgent    func(*WSConn) Agent
	//wg          sync.WaitGroup
	//ReadMaxMsgLen uint32
}

func (handler *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowd", 405)
		return
	}
	if r.URL.Path != "/digitalhuman-ws" {
		http.Error(w, "404 NOT FOUND", 404)
		return
	}

	// 升级为websocket协议
	conn, err := handler.upgrader.Upgrade(w, r, nil)
	if err != nil {
		wslog.Logger.Fatal(handler.ctx, "Upgrader error", logit.Error("error", err))
		http.Error(w, "Upgrade websocket failed!", 500)
		return
	}

	//conn.SetReadLimit(int64(handler.ReadMaxMsgLen))

	//handler.wg.Add(1)
	//defer handler.wg.Done()

	// 计算连接数
	handler.mutex.Lock()
	if handler.conns == nil {
		handler.mutex.Unlock()
		conn.Close()
		return
	}
	if len(handler.conns) >= handler.maxConnNum {
		handler.mutex.Unlock()
		conn.Close()
		wslog.Logger.Fatal(handler.ctx, "Too many connections!")
		return
	}
	// 链接相关操作
	ssid := utils.NewUUID()
	wsConn := newWSConn(conn, handler, handler.writeMsgCap, ssid)
	handler.conns[wsConn] = ssid
	handler.mutex.Unlock()
	agent := handler.newAgent(wsConn)
	agent.ReadMsg()
}

// Start 启动服务
func (server *WSServer) Start() {
	log.Println("Websocket server starting...")
	log.Println("Websocket address [" + server.Addr + "]")
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		panic(err)
	}
	if server.MaxConnNum <= 0 {
		server.MaxConnNum = 100
		log.Printf("Invalid MaxConnNum, reset to %v\n", server.MaxConnNum)
	}

	if server.WriteMsgCap <= 0 {
		server.WriteMsgCap = 100
		log.Printf("Invalid WriteMsgCap, reset to %v\n", server.WriteMsgCap)
	}

	//if server.ReadMaxMsgLen <= 0 {
	//	server.ReadMaxMsgLen = 4096
	//	log.Printf("Invalid ReadMaxMsgLen, reset to %v", server.ReadMaxMsgLen)
	//}

	if server.HTTPTimeout <= 0 {
		server.HTTPTimeout = 10 * time.Second
		log.Printf("Invalid HTTPTimeout, reset to %v\n", server.HTTPTimeout)
	}

	if server.CerFile != "" || server.KeyFile != "" {
		config := &tls.Config{}
		config.NextProtos = []string{"http/1.1"}

		var err error
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(server.CerFile, server.KeyFile)
		if err != nil {
			log.Fatal("Run TLS server failed, " + err.Error())
		}
		ln = tls.NewListener(ln, config)
	}

	server.ln = ln
	server.handler = &WSHandler{
		ctx:         server.Ctx,
		maxConnNum:  server.MaxConnNum,
		writeMsgCap: server.WriteMsgCap,
		upgrader: websocket.Upgrader{
			HandshakeTimeout: server.HTTPTimeout,
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		},
		conns:    make(WSConnPool),
		newAgent: server.NewAgent,
	}
	server.HTTPServer = &http.Server{
		Addr:           server.Addr,
		Handler:        server.handler,
		ReadTimeout:    server.HTTPTimeout,
		WriteTimeout:   server.HTTPTimeout,
		MaxHeaderBytes: 1024,
	}

	// 默认启服务方法
	//if err := server.HTTPServer.Serve(ln); err != nil {
	//	log.Println("websocket server stopped, error[" + err.Error() + "]")
	//}

	// 另外启一个进程是因为增加信号处理
	go func(failChan chan error) {
		if err := server.HTTPServer.Serve(ln); err != nil {
			failChan <- err
		}
	}(server.FailChan)
}

// Close 关闭服务
func (server *WSServer) Close() {
	server.ln.Close()
	for wsConn := range server.handler.conns {
		wsConn.Close()
	}
	server.handler.mutex.Lock()
	server.handler.conns = nil
	server.handler.mutex.Unlock()
	// 重启或者是杀掉主线程后，等待后续请求处理完
	//server.handler.wg.Wait()
}
