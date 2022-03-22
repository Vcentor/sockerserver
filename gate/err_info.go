// @Author: Vcentor
// @Date: 2020/11/24 2:28 下午

package gate

// JsonResponse json返回信息
type JsonResponse struct {
	RequestId string      `json:"requestId"`
	Action    string      `json:"action"`
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Body      interface{} `json:"body"`
}

const (
	ERR_PARSE_ROUTE    = 50000 // 解析路由失败
	ERR_REQUEST_PARAMS = 50001 // 非法的请求参数
)
