// @Author: Vcentor
// @Date: 2020/11/10 8:25 下午

package response

import (
	"socketserver/gate"
	"encoding/json"
)

// Resp 下行数据格式
type JsonResp struct {
	RequestId string      `json:"requestId"`
	Action    string      `json:"action"`
	Body      interface{} `json:"body"`
	Code      int         `json:"code"`
	Message   string      `json:"message"`
}

// NewResp 实例化返回数据
func NewResp(requestId, action string, body interface{}, code int, message string) *JsonResp {
	return &JsonResp{
		RequestId: requestId,
		Action:    action,
		Body:      body,
		Code:      code,
		Message:   message,
	}
}

func (j *JsonResp) JsonSend(a *gate.WSAgent) {
	b, _ := json.Marshal(j)
	a.Conn.WriteMsg(b)
}
