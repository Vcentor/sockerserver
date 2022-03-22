/**
 * @File: file.go
 * @Author: Vcentor
 * @Date: 2019/10/15 4:21 下午
 */
package utils

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
)

var (
	ErrEmptyCreatedPath = errors.New("created path cannot empty")
	ErrEmptyCreatedFile = errors.New("created file cannot empty")
)

// IsDirOrFile 判断是否是文件夹或文件
func IsDirOrFile(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true
	}
	return os.IsExist(err)
}

// GenDir 初始化各个文件的路径
func GenDir(s string) error {
	if b := IsDirOrFile(s); !b {
		err := os.MkdirAll(s, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

// GenFile 生成文件
func GenFile(filename string, b []byte) error {
	if len(b) == 0 {
		return errors.New("file content cannot be empty")
	}
	f, err := os.Create(filename)
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if err != nil {
		return err
	}
	return nil
}

// CreateDirs 创建资源目录
func CreateDirs(dir ...string) error {
	if len(dir) == 0 {
		return ErrEmptyCreatedPath
	}
	for _, d := range dir {
		dirname := path.Base(d)
		err := GenDir(d)
		if err != nil {
			return fmt.Errorf("Create %s dir failed!errmsg=%s", dirname, err)
		}
	}
	return nil
}

// CreateFiles 创建素材文件
func CreateFiles(resource map[string][]byte) error {
	if len(resource) == 0 {
		return ErrEmptyCreatedFile
	}
	for f, b := range resource {
		filename := path.Base(path.Dir(f)) + "/" + path.Base(f)
		err := GenFile(f, b)
		if err != nil {
			return fmt.Errorf("Create %s file failed!errmsg=%s", filename, err)
		}
	}
	return nil
}

// CopyFile 拷贝单个文件
func CopyFile(filename, dst string) error {
	fContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	dstFileInfo, err := os.Stat(dst)
	if err != nil {
		return err
	}
	if !dstFileInfo.IsDir() {
		return fmt.Errorf("CopyFile error, dst must be a dir")
	}
	dstFilename := path.Join(dst, path.Base(filename))
	dstFile, err := os.Create(dstFilename)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	if _, err := dstFile.Write(fContent); err != nil {
		return err
	}
	return nil
}

// CopyFiles 拷贝文件夹文件
func CopyFiles(src, dst string) error {
	srcFileInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	dstFileInfo, err := os.Stat(dst)
	if err != nil {
		return err
	}
	if !dstFileInfo.IsDir() || !srcFileInfo.IsDir() {
		return fmt.Errorf("CopyFiles error, src or dst must be a dir")
	}

	srcFileHandler, err := os.Open(src)
	if err != nil {
		return err
	}
	err = copyfiles(srcFileHandler, dst)
	if err != nil {
		return err
	}
	return nil
}

func copyfiles(file *os.File, dst string) error {
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		srcFileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range srcFileInfos {
			if fi.IsDir() {
				if err := os.MkdirAll(path.Join(dst, fi.Name()), 0775); err != nil {
					return err
				}
			}
			f, err := os.Open(path.Join(file.Name(), fi.Name()))
			if err != nil {
				return err
			}
			if err := copyfiles(f, path.Join(dst, fi.Name())); err != nil {
				return err
			}

		}
	} else {
		f, err := os.Create(dst)
		if err != nil {
			return err
		}
		_, err = io.Copy(f, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
