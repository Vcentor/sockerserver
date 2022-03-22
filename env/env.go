// @Author: Vcentor
// @Date: 2020/9/25 11:08 上午

package env

import (
	"os"
	"path"
	"path/filepath"
)

// RootPath 根目录
func RootPath() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic("Cannot get current dir: " + err.Error())
	}
	var dir string

	bindir := filepath.Dir(os.Args[0])
	if !filepath.IsAbs(bindir) {
		bindir = filepath.Join(pwd, bindir)
	}
	// 如果有和可执行文件平级的conf目录，则当前目录就是根目录
	// 这通常是直接在代码目录里go build然后直接执行生成的结果
	dir = filepath.Join(bindir, "conf")
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		dir = bindir
		return dir
	} else {
		panic("Cannot get conf dir by current dir")
	}
}

// ConfPath conf目录
func ConfPath() string {
	return path.Join(RootPath(), "conf")
}

// DataPath data目录
func DataPath() string {
	return path.Join(RootPath(), "data")
}

func LogPath() string {
	return path.Join(RootPath(), "log")
}

func LogicPath() string {
	return path.Join(RootPath(), "logic")
}
