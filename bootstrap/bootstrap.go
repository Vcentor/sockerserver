// @Author: Vcentor
// @Date: 2020/9/24 10:43 上午

package bootstrap

import (
	"context"
	"socketserver/gate"
	"socketserver/library/wslog"
	"socketserver/logic"
	"socketserver/processer"
)

// Bootstrap 程序启动入口
type Bootstrap struct {
	ctx context.Context
}

// NewBootstrap 初始化Bootstrap对象
func NewBootstrap(ctx context.Context) Bootstrap {
	return Bootstrap{ctx: ctx}
}

// Init 初始化一些配置选项
func (b Bootstrap) Init() Bootstrap {
	p := processer.NewJSONProcesser("requestId", "action", "body")
	wslog.Init(b.ctx)
	gate.Init(b.ctx, p)
	logic.Init(p)
	return b
}

// Run 运行服务
func (b Bootstrap) Run() {
	gate.Gateway.Run()
}
