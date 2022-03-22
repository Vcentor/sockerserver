// @Author: Vcentor
// @Date: 2020/9/23 3:08 下午

package main

import (
	"context"
	"socketserver/bootstrap"
)

func main() {
	// 当服务重启或者关闭时，等待其它请求完成，默认等待1s
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bootstrap.NewBootstrap(ctx).Init().Run()
}
