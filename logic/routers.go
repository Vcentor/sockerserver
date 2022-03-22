// @Author: Vcentor
// @Date: 2020/10/21 5:21 下午

package logic

import (
	"socketserver/logic/actions"
	"socketserver/processer"
)

// Init 初始化注册路由
func Init(processer processer.ProcesserOpt) {
	actions.Sample()
}
