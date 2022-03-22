// @Author: Vcentor
// @Date: 2020/9/24 10:58 上午

package gate

import (
	"context"
	"socketserver/env"
	"socketserver/network"
	"socketserver/processer"
	"github.com/BurntSushi/toml"
	"log"
	"os"
	"os/signal"
	"path"
	"time"
)

var Gateway = new(Gate)

// Gate 网关信息
type Gate struct {
	Ctx             context.Context
	ShutdownTimeout time.Duration `toml:"shutdown_timeout"`
	WSConf          WSConfOption  `toml:"ws_conf"`
	TCPConf         TPCConfOption `toml:"tcp_conf"`
}

// WSOption websocket服务配置选项
type WSConfOption struct {
	Processer   processer.ProcesserOpt
	IDC         string        `toml:"idc"`
	ListenAddr  string        `toml:"listen_addr"`
	MaxConnMum  int           `toml:"max_conn_num"`
	WriteMsgCap int           `toml:"write_msg_cap"`
	HTTPTimeout time.Duration `toml:"http_timeout"`
	CerFile     string        `toml:"cer_file"`
	KeyFile     string        `toml:"key_file"`
}

type TPCConfOption struct {
	ListenAddr string `toml:"listen_addr"`
	MaxConnNum int    `toml:"max_conn_num"`
	ChanCap    int    `toml:"chan_cap"`

	LenMsgLen    int    `toml:"len_msg_len"`
	MaxMsgLen    uint32 `toml:"max_msg_len"`
	MinMsgLen    uint32 `toml:"min_msg_len"`
	LittleEndian bool   `toml:"little_endian"`
}

// Init 初始化Gate
func Init(ctx context.Context, processer processer.ProcesserOpt) {
	serverConf := path.Join(env.ConfPath(), "server.toml")
	if _, err := toml.DecodeFile(serverConf, Gateway); err != nil {
		panic(err)
	}
	Gateway.Ctx = ctx
	Gateway.WSConf.Processer = processer
}

// Run 启动服务
func (gate *Gate) Run() {
	var wsserver *network.WSServer
	if gate.WSConf.ListenAddr != "" {
		wsserver = &network.WSServer{
			Ctx:         gate.Ctx,
			Addr:        gate.WSConf.ListenAddr,
			MaxConnNum:  gate.WSConf.MaxConnMum,
			WriteMsgCap: gate.WSConf.WriteMsgCap,
			HTTPTimeout: gate.WSConf.HTTPTimeout * time.Second,
			CerFile:     gate.WSConf.CerFile,
			KeyFile:     gate.WSConf.KeyFile,
			FailChan:    make(chan error),
			NewAgent: func(conn *network.WSConn) network.Agent {
				return &WSAgent{
					Conn: conn,
					Gate: gate,
				}
			},
		}
	}

	var tcpserver *network.TCPServer
	if gate.TCPConf.ListenAddr != "" {
		tcpserver = &network.TCPServer{
			Ctx:          gate.Ctx,
			ListenAddr:   gate.TCPConf.ListenAddr,
			MaxConnNum:   gate.TCPConf.MaxConnNum,
			ChanCap:      gate.TCPConf.ChanCap,
			LenMsgLen:    gate.TCPConf.LenMsgLen,
			MaxMsgLen:    gate.TCPConf.MaxMsgLen,
			MinMsgLen:    gate.TCPConf.MinMsgLen,
			LittleEndian: gate.TCPConf.LittleEndian,
			NewAgent: func(conn *network.TCPConn) network.Agent {
				return &TCPAgent{
					Conn: conn,
					Gate: gate,
				}
			},
		}
	}

	if wsserver != nil {
		wsserver.Start()
	}

	if tcpserver != nil {
		tcpserver.Start()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	select {
	case sig := <-c:
		log.Printf("Recieve signal %v\n", sig)
		// 关闭连接服务
		wsserver.Close()
		log.Println("Websocket server is closing...")

		tcpserver.Close()
		log.Println("TCP server is closing...")

		// 等待5s后关闭所有请求
		ctx, cancel := context.WithTimeout(gate.Ctx, gate.ShutdownTimeout*time.Second)
		defer cancel()
		wsserver.HTTPServer.Shutdown(ctx)
	case err := <-wsserver.FailChan:
		log.Fatal("websocket start failed, error[" + err.Error() + "]")
	}
}
