// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2020/7/13

package conf

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"icode.baidu.com/baidu/gdp/env"
)

// ErrNoParser 没有找到解析函数
var ErrNoParser = fmt.Errorf("no parser found")

// Conf 配置解析定义
type Conf interface {
	// 读取并解析配置文件
	// confName 支持相对路径和绝对路径
	Parse(confName string, obj interface{}) error

	// 解析bytes内容
	ParseBytes(fileExt string, content []byte, obj interface{}) error

	// 配置文件是否存在
	Exists(confName string) bool

	// 注册一个指定后缀的配置的parser
	// 如要添加 .ini 文件的支持，可在此注册对应的解析函数即可
	RegisterParserFunc(fileExt string, fn ParserFunc) error

	// 注册一个在解析前执行辅助回调方法
	// 先注册的先执行，不能重复
	RegisterBeforeFunc(name string, fn BeforeFunc) error

	// 配置的环境信息
	Env() env.AppEnv
}

// New 创建一个新的配置解析实例
// 返回的实例是没有注册任何解析能力的
func New(e env.AppEnv) Conf {
	conf := &conf{
		parsers: map[string]ParserFunc{},
		env:     e,
	}

	return conf
}

// NewDefault 创建一个新的配置解析实例
// 会注册默认的配置解析方法和辅助方法
func NewDefault(e env.AppEnv) Conf {
	conf := New(e)
	for name, fn := range DefaultParserFuncs {
		_ = conf.RegisterParserFunc(name, fn)
	}

	for _, h := range defaultHelpers {
		if err := conf.RegisterBeforeFunc(h.Name, h.Func); err != nil {
			panic(fmt.Sprintf("RegisterHelper(%q) err=%s", h.Name, err))
		}
	}
	return conf
}

type conf struct {
	env     env.AppEnv
	parsers map[string]ParserFunc
	helpers []*beforeHelper
}

func (c *conf) Parse(confName string, obj interface{}) (err error) {
	confAbsPath := c.confFileRealPath(confName)
	return c.parseByAbsPath(confAbsPath, obj)
}

var relPathPre = "." + string(filepath.Separator)

func (c *conf) confFileRealPath(confName string) string {
	if filepath.IsAbs(confName) ||
		strings.HasPrefix(confName, "./") ||
		strings.HasPrefix(confName, relPathPre) {
		return confName
	}
	return filepath.Join(c.Env().ConfDir(), confName)
}

func (c *conf) parseByAbsPath(confAbsPath string, obj interface{}) (err error) {
	if len(c.parsers) == 0 {
		return ErrNoParser
	}

	return c.readConfDirect(confAbsPath, obj)
}

func (c *conf) readConfDirect(confPath string, obj interface{}) error {
	content, errIO := ioutil.ReadFile(confPath)
	if errIO != nil {
		return errIO
	}
	fileExt := filepath.Ext(confPath)
	return c.ParseBytes(fileExt, content, obj)
}

func (c *conf) Env() env.AppEnv {
	if c.env == nil {
		return env.Default
	}
	return c.env
}

func (c *conf) ParseBytes(fileExt string, content []byte, obj interface{}) error {
	parserFn, hasParser := c.parsers[fileExt]
	if fileExt == "" || !hasParser {
		return fmt.Errorf("%w, fileExt %q is not supported yet", ErrNoParser, fileExt)
	}

	contentNew, errHelper := c.executeBeforeHelpers(content, c.helpers)

	if errHelper != nil {
		return fmt.Errorf("%w, content=\n%s", errHelper, string(contentNew))
	}

	if errParser := parserFn(contentNew, obj); errParser != nil {
		return fmt.Errorf("%w, content=\n%s", errParser, string(contentNew))
	}
	return nil
}

// executeBeforeHelpers 执行
func (c *conf) executeBeforeHelpers(input []byte, helpers []*beforeHelper) (output []byte, err error) {
	if len(helpers) == 0 {
		return input, nil
	}
	output = input
	for _, helper := range helpers {
		output, err = helper.Func(c, output)
		if err != nil {
			return nil, fmt.Errorf("beforeHelper=%q has error:%w", helper.Name, err)
		}
	}
	return output, err
}

func (c *conf) Exists(confName string) bool {
	info, err := os.Stat(c.confFileRealPath(confName))
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func (c *conf) RegisterParserFunc(fileExt string, fn ParserFunc) error {
	if _, has := c.parsers[fileExt]; has {
		return fmt.Errorf("parser=%q already exists", fileExt)
	}
	c.parsers[fileExt] = fn
	return nil
}

func (c *conf) RegisterBeforeFunc(name string, fn BeforeFunc) error {
	if name == "" {
		return fmt.Errorf("name is empty, not allow")
	}

	for _, h1 := range c.helpers {
		if name == h1.Name {
			return fmt.Errorf("beforeHelper=%q already exists", name)
		}
	}
	c.helpers = append(c.helpers, newBeforeHelper(name, fn))
	return nil
}

var _ Conf = (*conf)(nil)
