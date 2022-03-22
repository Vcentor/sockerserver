// @Author: Vcentor
// @Date: 2020/10/20 11:39 上午

package processer

// ProcesserOpt 处理器接口
type ProcesserOpt interface {
	Unmarshal([]byte) (Processer, error)
	Route(Processer, interface{}) error
	RegisterRouter(string, func(string, []byte, interface{}))
}

// Processer 处理器对象
type Processer struct {
	RequestID string
	Action    string
	Body      []byte
}
