#初始化项目目录变量
HOMEDIR := $(shell pwd)
OUTDIR  := $(HOMEDIR)/output
#初始化命令变量
GOROOT  := $(HOMEDIR)/../../../baidu/go-env/go1-16-linux-amd64/
GO      := $(GOROOT)/bin/go
GOPATH  := $(shell $(GO) env GOPATH)
GOMOD   := $(GO) mod
GOBUILD := $(GO) build
GOTEST  := $(GO) test
GOPKGS  := $$($(GO) list ./...| grep -vE "vendor")
#执行编译，可使用命令 make 或 make all 执行, 顺序执行prepare -> compile -> test -> package 几个阶段
all: prepare compile test package
# prepare阶段, 使用bcloud下载非Go依赖，使用GOD下载Go依赖, 可单独执行命令: make prepare
prepare: prepare-dep
prepare-dep:
	git config --global http.sslVerify false #设置git， 保证github mirror能够下载
	bcloud local -U #下载非Go依赖，依赖之前的BCLOUD文件
set-env:
	$(GO) env -w GONOPROXY=\*.baidu.com\*
	$(GO) env -w GOPROXY=https://goproxy.baidu-int.com/
	$(GO) env -w GONOSUMDB=\*
#complile阶段，执行编译命令,可单独执行命令: make compile
compile:build
build: set-env
	$(GOMOD) tidy #下载Go依赖
	mkdir $(OUTDIR)
#test阶段，进行单元测试， 可单独执行命令: make test
test: test-case
test-case: set-env
	$(GOTEST) -v -race -cover $(GOPKGS)
#与覆盖率平台打通，输出测试结果到文件中
#@$(GOTEST) -v -json -coverprofile=coverage.out $(GOPKGS) > testlog.out
#package阶段，对编译产出进行打包，输出到output目录, 可单独执行命令: make package
package: package-bin
package-bin:
	cp README.md  $(OUTDIR)/

#clean阶段，清除过程中的输出, 可单独执行命令: make clean
clean:
	rm -rf $(OUTDIR)
# avoid filename conflict and speed up build
.PHONY: all prepare compile test package  clean build
