// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2020/7/10

package env

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	// DefaultIDC 默认idc的值
	// idc 推荐可选值  如 jx,gz等
	DefaultIDC = "test"

	// DefaultAppName 默认的app名称
	DefaultAppName = "unknown"

	// DefaultRunMode 测试默认运行等级
	DefaultRunMode = RunModeRelease
)

// 可以依据不同的运行等级来开启不同的调试功能、接口
const (
	// RunModeDebug 调试
	RunModeDebug = "debug"

	// RunModeTest 测试
	RunModeTest = "test"

	// RunModeRelease 线上发布
	RunModeRelease = "release"
)

// Option 具体的环境信息
//
// 所有的选项都是可选的
type Option struct {
	// AppName 应用名称
	AppName string

	// IDC 所在idc
	IDC string

	// RunMode 运行模式
	RunMode string

	// RootDir 应用根目录地址
	// 若为空，将通过自动推断的方式获取
	RootDir string

	// DataDir 应用数据根目录地址
	// 默认为 RootDir+"/data/"
	DataDir string

	// LogDir 应用日志根目录地址
	// 默认为 RootDir+"/log/"
	LogDir string

	// ConfDir 应用配置文件根目录地址
	// 默认为RootDir+"/conf/"
	ConfDir string
}

// String 序列化，方便查看
// 目前输出的是一个json
func (opt Option) String() string {
	format := `{"AppName":%q,"IDC":%q,"RootDir":%q,"DataDir":%q,"LogDir":%q,"ConfDir":%q,"RunMode":%q}`
	return fmt.Sprintf(format, opt.AppName, opt.IDC, opt.RootDir, opt.DataDir, opt.LogDir, opt.ConfDir, opt.RunMode)
}

// Merge 合并
// 传入的Option不为空则合并，否则使用老的值
func (opt Option) Merge(b Option) Option {
	return Option{
		AppName: chooseBFirst(opt.AppName, b.AppName),
		IDC:     chooseBFirst(opt.IDC, b.IDC),
		RunMode: chooseBFirst(opt.RunMode, b.RunMode),
		RootDir: chooseBFirst(opt.RootDir, b.RootDir),
		DataDir: chooseBFirst(opt.DataDir, b.DataDir),
		LogDir:  chooseBFirst(opt.LogDir, b.LogDir),
		ConfDir: chooseBFirst(opt.ConfDir, b.ConfDir),
	}
}

func chooseBFirst(a, b string) string {
	if b != "" {
		return b
	}
	return a
}

// AppEnv 应用环境信息完整的接口定义
type AppEnv interface {
	// 应用名称
	AppNameEnv

	// idc信息
	IDCEnv

	// 应用根目录
	RootDirEnv

	// 应用配置文件根目录
	ConfDirEnv

	// 应用数据文件根目录
	DataDirEnv

	// 应用日志文件更目录
	LogDirEnv

	// 应用运行情况
	RunModeEnv

	// 获取当前环境的选项详情
	Options() Option

	// 复制一个新的env对象，并将传入的Option merge进去
	CloneWithOption(opt Option) AppEnv
}

// RootDirEnv 应用根目录环境信息
type RootDirEnv interface {
	RootDir() string
}

// ConfDirEnv 配置环境信息
type ConfDirEnv interface {
	ConfDir() string
}

// DataDirEnv 数据目录环境信息
type DataDirEnv interface {
	DataDir() string
}

// LogDirEnv 日志目录环境信息
type LogDirEnv interface {
	LogDir() string
}

// IDCEnv idc的信息读写接口
type IDCEnv interface {
	IDC() string
}

// AppNameEnv 应用名称
type AppNameEnv interface {
	AppName() string
}

// RunModeEnv 运行模式/等级
type RunModeEnv interface {
	RunMode() string
}

// New 创建新的应用环境
func New(opt Option) AppEnv {
	env := &appEnv{}

	if opt.AppName != "" {
		env.setAppName(opt.AppName)
	}

	if opt.IDC != "" {
		env.setIDC(opt.IDC)
	}

	if opt.RunMode != "" {
		env.setRunMode(opt.RunMode)
	}

	if opt.RootDir != "" {
		env.setRootDir(opt.RootDir)
	}
	if opt.ConfDir != "" {
		env.setConfDir(opt.ConfDir)
	}
	if opt.DataDir != "" {
		env.setDataDir(opt.DataDir)
	}

	if opt.LogDir != "" {
		env.setLogDir(opt.LogDir)
	}
	return env
}

type appEnv struct {
	rootDir string
	dataDir string
	confDir string
	logDir  string

	appName string
	idc     string

	runMode string
}

func (a *appEnv) AppName() string {
	if a.appName != "" {
		return a.appName
	}
	return DefaultAppName
}

func (a *appEnv) setAppName(name string) {
	setValue(&a.appName, name, "AppName")
}

func (a *appEnv) IDC() string {
	if a.idc == "" {
		return DefaultIDC
	}
	return a.idc
}

func (a *appEnv) setIDC(idc string) {
	setValue(&a.idc, idc, "IDC")
}

func (a *appEnv) RunMode() string {
	if a.runMode != "" {
		return a.runMode
	}
	return DefaultRunMode
}

func (a *appEnv) setRunMode(mod string) {
	setValue(&a.runMode, mod, "RunMode")
}

func (a *appEnv) RootDir() string {
	if a.rootDir != "" {
		return a.rootDir
	}
	return AutoDetectAppRootDir()
}

func (a *appEnv) setRootDir(dir string) {
	setValue(&a.rootDir, dir, "RootDir")
}

func (a *appEnv) DataDir() string {
	return a.chooseDir(a.dataDir, "data")
}

func (a *appEnv) setDataDir(dir string) {
	setValue(&a.dataDir, dir, "DataDir")
}

func (a *appEnv) LogDir() string {
	return a.chooseDir(a.logDir, "log")
}

func (a *appEnv) setLogDir(dir string) {
	setValue(&a.logDir, dir, "LogDir")
}

func (a *appEnv) ConfDir() string {
	return a.chooseDir(a.confDir, "conf")
}

func (a *appEnv) setConfDir(dir string) {
	setValue(&a.confDir, dir, "ConfDir")
}

func (a *appEnv) chooseDir(dir string, subDirName string) string {
	if dir != "" {
		return dir
	}
	return filepath.Join(a.RootDir(), subDirName)
}

func (a *appEnv) Options() Option {
	return Option{
		AppName: a.AppName(),
		IDC:     a.IDC(),
		RunMode: a.RunMode(),

		RootDir: a.RootDir(),
		DataDir: a.DataDir(),
		LogDir:  a.LogDir(),
		ConfDir: a.ConfDir(),
	}
}

func (a *appEnv) CloneWithOption(opt Option) AppEnv {
	opts := a.Options().Merge(opt)
	return New(opts)
}

var _ AppEnv = (*appEnv)(nil)

func setValue(addr *string, value string, fieldName string) {
	*addr = value
	_ = log.Output(2, fmt.Sprintf("[env] set %q=%q\n", fieldName, value))
}

// AutoDetectAppRootDir 自动获取应用根目录
// 定义为变量，这样若默认实现不满足，可进行替换
var AutoDetectAppRootDir = autoDetect

func autoDetect() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	names := []string{
		"go.mod",
		filepath.Join("conf", "app.toml"),
	}
	dir, err := findDirMatch(wd, names)
	if err == nil {
		return dir
	}
	return wd
}

var errNotFound = fmt.Errorf("cannot found")

// findDirMatch 在指定目录下，向其父目录查找对应的文件是否存在
// 若存在，则返回匹配到的路径
func findDirMatch(baseDir string, fileNames []string) (dir string, err error) {
	currentDir := baseDir
	for i := 0; i < 20; i++ {
		for _, fileName := range fileNames {
			depsPath := filepath.Join(currentDir, fileName)
			if _, err := os.Stat(depsPath); !os.IsNotExist(err) {
				return currentDir, nil
			}
		}

		currentDir = filepath.Dir(currentDir)

		if currentDir == "." {
			break
		}
	}
	return "", errNotFound
}
