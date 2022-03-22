// @Author: Vcentor
// @Date: 2020/10/21 3:27 下午

package gate

import (
	"socketserver/library/wslog"
	"socketserver/network"
	"encoding/base64"
	"encoding/json"
	"icode.baidu.com/baidu/gdp/logit"
)

// websocket读数据的类型
const (
	TEXT_MESSAGE   = 1
	BINARY_MESSAGE = 2
)

// GlobalData 保存端上全局数据
type GlobalData struct {
	DataType string // 区分同一链接的不同公共数据，最好以action命名，充分解耦
	Data     []byte
}

var _ network.Agent = (*WSAgent)(nil)

// WSAgent 网关接口代理
type WSAgent struct {
	Conn       *network.WSConn
	Gate       *Gate
	AsrConn    *network.WSClient
	GlobalData *GlobalData
}

// ReadMsg 读信息
func (a *WSAgent) ReadMsg() {
	for {
		data, messageType, err := a.Conn.ReadMsg()
		if err != nil {
			wslog.Logger.Notice(a.Gate.Ctx, "Read message failed", logit.Error("error", err))
			goto CLOSE
		}
		switch messageType {
		case BINARY_MESSAGE:
			requestId := a.Conn.GetSessionID()
			d := []byte(`{
				"action":"PROCESS_PCM",
				"requestId": "` + requestId + `",
				"body": "` + base64.StdEncoding.EncodeToString(data) + `"
			}`)
			msg, err := a.Gate.WSConf.Processer.Unmarshal(d)
			if err != nil {
				wslog.Logger.Fatal(a.Gate.Ctx, "Unmarshal binary message failed", logit.Error("error", err))
				// switch层面的break 不断开连接
				break
			}
			a.Gate.WSConf.Processer.Route(msg, a)
		case TEXT_MESSAGE:
			wslog.Logger.Notice(a.Gate.Ctx, "text message data", logit.String("data", string(data)))
			msg, err := a.Gate.WSConf.Processer.Unmarshal(data)
			if err != nil {
				wslog.Logger.Fatal(a.Gate.Ctx, "Unmarshal text message failed", logit.Error("error", err))
				var resp = JsonResponse{
					Code:    ERR_REQUEST_PARAMS,
					Message: "Illegal request params",
				}
				b, _ := json.Marshal(&resp)
				a.Conn.WriteMsg(b)
				goto CLOSE
			}

			if err := a.Gate.WSConf.Processer.Route(msg, a); err != nil {
				wslog.Logger.Fatal(a.Gate.Ctx, "Route failed", logit.Error("error", err))
				var resp = JsonResponse{
					RequestId: msg.RequestID,
					Action:    "UNKOWN",
					Code:      ERR_PARSE_ROUTE,
					Message:   "Illegal action!",
					Body:      nil,
				}
				b, _ := json.Marshal(&resp)
				a.Conn.WriteMsg(b)
				goto CLOSE
			}
		}
	}
CLOSE:
	a.Conn.Close()
}
