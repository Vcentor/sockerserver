// @Author: Vcentor
// @Date: 2020/10/14 1:38 下午

package wslog

import (
	"bytes"
	"context"
	"socketserver/env"
	"socketserver/library/utils"
	"time"

	"icode.baidu.com/baidu/gdp/logit"
)

var Logger logit.Logger

func Init(ctx context.Context) {
	var err error
	c := &logit.Config{
		FileName:   env.LogPath() + "/service/service.log",
		RotateRule: "1hour", // 每1小时产生一个新的文件
		MaxFileNum: 48,      // 保留最新的48个日志文件
		// PrefixFunc:    logit.DefaultPrefixFunc, // 日志每行的前缀部分
		PrefixFunc: func(ctx context.Context, level logit.Level, callDepth int) []byte {
			var buf bytes.Buffer
			buf.WriteString(level.String() + " ")
			buf.WriteString(utils.TimeStamp2Datetime(time.Now()))
			buf.Write([]byte(" "))
			return buf.Bytes()
		},
		BufferSize:    1024, // 若为0会使用默认值 4096，-1 是禁用(值0)
		WriterTimeout: 0,
	}

	Logger, err = logit.NewLogger(ctx, logit.OptConfig(c))
	if err != nil {
		panic(err)
	}
}
