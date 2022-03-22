// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2020/8/20

package logit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"icode.baidu.com/baidu/gdp/conf"
	"icode.baidu.com/baidu/gdp/extension/pool"
	"icode.baidu.com/baidu/gdp/extension/writer"
)

// ConfigDispatch 日志分发规则的配置部分
type ConfigDispatch struct {
	// 文件后缀，可以为空，若为空，则不添加后缀，也就是默认的notice日志
	// 如 为空时，日志文件是 service.log.2020081011
	// 如 为 '.wf' 是，日志文件是 service.log.wf.2020081011
	FileSuffix string

	// 日志等级，支持配置多个，只有配置的，才会打印到日志文件里去
	// 对于配置文件，需要配置为字符串，详见下面的 configDispatchConf 类型
	Levels []Level
}

// configDispatchConf 配置文件中的ConfigDispatch格式
type configDispatchConf struct {
	FileSuffix string
	Levels     []string // 在配置的时候，可以直接使用"NOTICE"这样的字符串表示
}

func (fc *configDispatchConf) ToFileOptionDispatch() (*ConfigDispatch, error) {
	fd := &ConfigDispatch{
		FileSuffix: fc.FileSuffix,
	}
	for _, ls := range fc.Levels {
		l, err := ParseLevel(ls)
		if err != nil {
			return nil, err
		}
		fd.Levels = append(fd.Levels, l)
	}
	return fd, nil
}

// PrefixFunc 日志内容前缀获取方法
type PrefixFunc func(ctx context.Context, level Level, callDepth int) []byte

// BeforeOutputFunc 在logger 的 Output 执行前执行
type BeforeOutputFunc func(ctx context.Context, enc FieldEncoder, level Level, callDepth int)

// Config 文件logger配置选项
// 配置示例可以看 _examples/conf/log 目录下的 ral.toml 和 ral-worker.toml
type Config struct {
	// 日志文件名
	// 如  xxx/log/abc.log
	FileName string

	Dispatch []*ConfigDispatch

	// 文件切分规则，如 1hour,1day,no,默认为1hour
	// 如 1hour 会将文件添加类似后缀 .2020072413
	// no 则不切分
	// 详细规则情况：
	// http://icode.baidu.com/repos/baidu/gdp/extension/blob/master:writer/rotate_producer.go
	RotateRule string

	// 保留最多日志文件数，默认为48，若为-1,则不会清理
	// 对于 FileName 所在目录下的 以FileName为前缀的文件将自动进行清理
	// 清理后剩余文件数量，清理周期同 RotateRule
	MaxFileNum int

	// 每行日志的前缀获取方法，
	// 若为nil，会使用默认的 DefaultPrefixFunc
	PrefixFunc PrefixFunc `json:"-"`

	// 日志前缀的方法名，用于查询 PrefixFunc
	// 若查询不到，config parser 会失败
	Prefix string

	// 在调用 logger.Output 方法时执行
	BeforeOutputFunc BeforeOutputFunc `json:"-"`

	// 用于查询 BeforeOutputFunc
	// 若查询不到，config parser 会失败
	BeforeOutput string

	// 日志内容待写缓冲队列大小
	// 若<0, 则是同步的
	// 若为0，则使用默认值4096
	BufferSize int

	// 日志进入待写队列超时时间，毫秒
	// 默认为0，不超时，若出现落盘慢的时候，调用写日志的地方会出现同步等待
	WriterTimeout int

	// 日志落盘刷新间隔，毫秒
	// 若<=0，使用默认值1000
	FlushDuration int

	// 日志编码的对象池名称
	// 若为空，会使用 default_text
	// 可以使用 RegisterEncoderPool 方法注册自定义encoder
	EncoderPool string

	encoderPool EncoderPool

	parsed bool

	// Blogger 专用
	binaryEncoder func(msg interface{}) ([]byte, error)

	writer io.WriteCloser

	// 用于监听文件变化的事件
	eventListen []func(w io.Writer)

	err error
}

func (cfg *Config) addEventListen(fn func(w io.Writer)) {
	cfg.eventListen = append(cfg.eventListen, fn)
}

// DefaultMaxFileNum 默认文件保留数
var DefaultMaxFileNum = 48

// parser 配置项解析
func (cfg *Config) parser() error {
	if cfg.parsed {
		return nil
	}

	if cfg.writer == nil {
		// 以下内容是创建一个writer所需要的配置
		if cfg.WriterTimeout < 0 {
			return fmt.Errorf(" WriterTimeout min value is 0, now is %d", cfg.WriterTimeout)
		}

		if cfg.FileName == "" {
			return fmt.Errorf(" FileName is required")
		}

		// 默认1小时切分一个新文件
		if cfg.RotateRule == "" {
			cfg.RotateRule = "1hour"
		}

		if cfg.BufferSize == 0 {
			cfg.BufferSize = 4096
		} else if cfg.BufferSize < 0 {
			cfg.BufferSize = 0
		}

		if cfg.MaxFileNum == 0 {
			cfg.MaxFileNum = DefaultMaxFileNum
		}
	}

	{
		// 刷新间隔，毫秒，默认值1000ms
		if cfg.FlushDuration <= 0 {
			cfg.FlushDuration = 1000
		}
	}

	{
		if len(cfg.Dispatch) == 0 {
			cfg.Dispatch = []*ConfigDispatch{
				{
					FileSuffix: "", // 为空是 默认的 notice日志
					Levels:     []Level{DebugLevel, TraceLevel, NoticeLevel},
				},
				{
					FileSuffix: ".wf",
					Levels:     []Level{WarningLevel, ErrorLevel, FatalLevel},
				},
			}
		}
		// 对FileSuffix 判断是否重复
		fileSuffixes := make(map[string]int, len(cfg.Dispatch))
		for idx, item := range cfg.Dispatch {
			if lastID, has := fileSuffixes[item.FileSuffix]; has {
				return fmt.Errorf("option: Dispatch.%d.FileSuffix=%q has duplicate with %d's", idx, item.FileSuffix, lastID)
			}
			fileSuffixes[item.FileSuffix] = idx
		}
	}

	if cfg.EncoderPool == "" {
		cfg.EncoderPool = encoderPoolNameDefaultText
	}

	cfg.encoderPool = GetEncoderPool(cfg.EncoderPool)

	if cfg.encoderPool == nil {
		return fmt.Errorf(" EncoderPool=%q not found", cfg.EncoderPool)
	}

	if cfg.PrefixFunc == nil {
		if cfg.Prefix != "" {
			if fn := GetPrefixFunc(cfg.Prefix); fn != nil {
				cfg.PrefixFunc = fn
			} else {
				return fmt.Errorf("not found Prefix=%q", cfg.Prefix)
			}
		} else {
			cfg.PrefixFunc = DefaultPrefixFunc
		}
	}

	if cfg.BeforeOutputFunc == nil && cfg.BeforeOutput != "" {
		fn := GetBeforeOutputFunc(cfg.BeforeOutput)
		if fn == nil {
			return fmt.Errorf(" BeforeOutput=%q not found", cfg.BeforeOutput)
		}
		cfg.BeforeOutputFunc = fn
	}

	cfg.parsed = true

	return nil
}

// nopLog 是否不需要打印日志
func (cfg *Config) nopLog() bool {
	if len(cfg.Dispatch) == 0 {
		return true
	}
	total := 0
	for _, item := range cfg.Dispatch {
		total += len(item.Levels)
	}
	return total == 0
}

func (cfg *Config) fireWriterChange(w io.Writer) {
	for i := 0; i < len(cfg.eventListen); i++ {
		cfg.eventListen[i](w)
	}
}

func (cfg *Config) getWriter() (io.WriteCloser, error) {
	if cfg.writer != nil {
		return cfg.writer, nil
	}

	rp, err := writer.NewSimpleRotateProducer(cfg.RotateRule, cfg.FileName)
	if err != nil {
		return nil, err
	}

	writerOption := &writer.RotateOption{
		FileProducer:  rp,
		FlushDuration: time.Duration(cfg.FlushDuration) * time.Millisecond,
		CheckDuration: 1 * time.Second,
		MaxFileNum:    cfg.MaxFileNum,
		NewWriter: func(w io.Writer) writer.RotateOptionWriter {
			cfg.fireWriterChange(w)
			return writer.NewLineBuffer(w)
		},
	}

	w, errRw := writer.NewRotate(writerOption)
	if errRw != nil {
		return nil, errRw
	}

	awc := writer.NewAsync(cfg.BufferSize, time.Millisecond*time.Duration(cfg.WriterTimeout), w)
	return awc, nil
}

func (cfg *Config) string() string {
	bf, err := json.Marshal(cfg)
	if err != nil {
		return err.Error()
	}
	return string(bf)
}

// LoadConfig 通过解析配置文件来来获取一个 Config
//
func LoadConfig(confName string) (*Config, error) {
	var config map[string]interface{}
	if err := conf.Parse(confName, &config); err != nil {
		return nil, err
	}
	if obj, has := config["Dispatch"]; has {
		bf, err := json.Marshal(obj)
		if err != nil {
			return nil, fmt.Errorf("parser Dispatch failed1:%w", err)
		}
		var ds []*configDispatchConf
		if err := json.Unmarshal(bf, &ds); err != nil {
			return nil, fmt.Errorf("parser Dispatch failed2:%w", err)
		}
		if len(ds) == 0 {
			delete(config, "Dispatch")
		} else {
			var do []*ConfigDispatch
			for _, item := range ds {
				d, err := item.ToFileOptionDispatch()
				if err != nil {
					return nil, err
				}
				do = append(do, d)
			}
			config["Dispatch"] = do
		}

	}

	bf, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	var opt *Config
	if err := json.Unmarshal(bf, &opt); err != nil {
		return nil, err
	}

	if err := opt.parser(); err != nil {
		return nil, err
	}

	return opt, nil
}

var prefixFuncPool = map[string]PrefixFunc{
	"default": DefaultPrefixFunc,
	"no":      prefixFuncNo,
	"default_nano": func(ctx context.Context, level Level, callDepth int) []byte {
		return samplePrefix(ctx, level, callDepth+1, "2006-01-02 15:04:05.999999999")
	},
}

func prefixFuncNo(ctx context.Context, level Level, callDepth int) []byte {
	return nil
}

// RegisterPrefixFunc 注册日志前缀计算方法
func RegisterPrefixFunc(name string, fn PrefixFunc) error {
	if _, has := prefixFuncPool[name]; has {
		return fmt.Errorf("name=%q already exists", name)
	}
	prefixFuncPool[name] = fn
	return nil
}

// GetPrefixFunc 获取注册的日志前缀计算方法
func GetPrefixFunc(name string) PrefixFunc {
	return prefixFuncPool[name]
}

// DefaultPrefixFunc 默认的日志前缀
// 	如 NOTICE: 2020-06-09 16:05:53 gdp/apiserver@v1.0.202003311755/context/log.go:95
var DefaultPrefixFunc = func(ctx context.Context, level Level, callDepth int) []byte {
	return samplePrefix(ctx, level, callDepth+1, "2006-01-02 15:04:05")
}

var bpPre = pool.NewBytesPool()

func samplePrefix(ctx context.Context, level Level, callDepth int, timeFormat string) []byte {
	buf := bpPre.Get()
	defer bpPre.Put(buf)

	buf.WriteString(level.String())
	buf.Write([]byte(": "))
	buf.WriteString(now().Format(timeFormat))
	buf.Write([]byte(" "))
	caller := FindField(ctx, callerKey)
	var callerPath string
	if caller != nil {
		if val, ok := caller.Value().(string); ok {
			callerPath = val
		}
	}
	if len(callerPath) == 0 {
		callerPath = callerWithSkip(callDepth + 1)
	}
	buf.WriteString(callerPath)
	buf.Write([]byte(" "))

	data := buf.Bytes()
	ret := make([]byte, len(data))
	copy(ret, data)
	return ret
}

var beforeOutputFuncPool = map[string]BeforeOutputFunc{
	// default ：什么都不做，默认的 DefaultPrefixFunc 已经将level等字段输出在日志的前部
	// 所以这个default应该什么都不做
	"default": func(ctx context.Context, enc FieldEncoder, level Level, callDepth int) {
	},

	// to_body : 将字段写入到 log 的 body 中去
	// 如果有些场景，比如输出json格式的日志，不需要日志前缀
	// 那么，可以设置 Prefix="no" , BeforeOutput="toBody"
	"to_body": beforeOutputFuncToBody,
}

// RegisterBeforeOutputFunc 注册 BeforeOutputFunc
func RegisterBeforeOutputFunc(name string, fn BeforeOutputFunc) error {
	if _, has := beforeOutputFuncPool[name]; has {
		return fmt.Errorf("name=%q already exists", name)
	}
	beforeOutputFuncPool[name] = fn
	return nil
}

// GetBeforeOutputFunc 获取已注册的 BeforeOutputFunc
func GetBeforeOutputFunc(name string) BeforeOutputFunc {
	return beforeOutputFuncPool[name]
}

func beforeOutputFuncToBody(ctx context.Context, enc FieldEncoder, level Level, callDepth int) {
	enc.AddString("level", level.String())
	enc.AddTime("logtime", now())

	caller := FindField(ctx, callerKey)
	var callerPath string
	if caller != nil {
		if val, ok := caller.Value().(string); ok {
			callerPath = val
		}
	}
	if len(callerPath) == 0 {
		callerPath = callerWithSkip(callDepth + 1)
	}
	enc.AddString("callerPath", callerPath)
}
