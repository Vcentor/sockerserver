// @Author: Vcentor
// @Date: 2020/10/20 11:32 上午

package processer

import (
	"encoding/json"
	"errors"
	"log"
)

// JSONProcesser json解析器
type JSONProcesser struct {
	RequestIDField string
	ActionField    string
	BodyField      string
	router         map[string]func(string, []byte, interface{})
}

// NewJSONProcesser 初始化processer
func NewJSONProcesser(requestIDField, actionField, bodyField string) *JSONProcesser {
	return &JSONProcesser{
		RequestIDField: requestIDField,
		ActionField:    actionField,
		BodyField:      bodyField,
		router:         make(map[string]func(string, []byte, interface{})),
	}
}

// Unmarshal 解析接收数据
func (j *JSONProcesser) Unmarshal(data []byte) (Processer, error) {
	var (
		m map[string]string
		p Processer
	)
	if err := json.Unmarshal(data, &m); err != nil {
		return p, err
	}
	reqeustID, ok := m[j.RequestIDField]
	if !ok {
		return p, errors.New("ReqeustIdField cannot parse")
	}
	p.RequestID = reqeustID
	action, ok := m[j.ActionField]
	if !ok {
		return p, errors.New("ActionField cannot parse")
	}
	p.Action = action
	p.Body = make([]byte, 0)
	body, ok := m[j.BodyField]
	// 这样写是因为端上发送心跳时，没有body字段
	if ok {
		p.Body = []byte(body)
	}
	return p, nil
}

// Route 路由
func (j *JSONProcesser) Route(p Processer, agent interface{}) error {
	handle, ok := j.router[p.Action]
	if !ok {
		return errors.New("route not register, requestId=" + p.RequestID + ", action=" + p.Action)
	}
	handle(p.RequestID, p.Body, agent)
	return nil
}

// RegisterRouter 注册路由
func (j *JSONProcesser) RegisterRouter(action string, handler func(string, []byte, interface{})) {
	if _, ok := j.router[action]; ok {
		log.Println(action + " already register")
		return
	}
	j.router[action] = handler
}
