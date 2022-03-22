// Author: Vcentor
// Date: 2022/3/22 3:12 下午
// desc:

package gate

import (
	"socketserver/library/wslog"
	"socketserver/network"
	"fmt"
	"icode.baidu.com/baidu/gdp/logit"
	"io"
)

var _ network.Agent = (*TCPAgent)(nil)

type TCPAgent struct {
	Conn *network.TCPConn
	Gate *Gate
}

func (a *TCPAgent) ReadMsg() {
	for {
		data, err := a.Conn.ReadMsg()
		if err != nil {
			if err == io.EOF {
				wslog.Logger.Notice(a.Gate.Ctx, "TCPServer connect closed", logit.Error("error", err))
			} else {
				wslog.Logger.Fatal(a.Gate.Ctx, "TCPServer read message  error", logit.Error("error", err))
			}

			goto CLOSE
		}

		wslog.Logger.Debug(a.Gate.Ctx, "read message", logit.String("info", string(data)))

		fmt.Println(a.Conn.WriteMsg(data))
		fmt.Println(a.Conn.GetSessionPool())
	}
CLOSE:
	a.Conn.Close()
}
