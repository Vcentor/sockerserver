#初始化项目目录变量
HOMEDIR := $(shell pwd)
OUTDIR  := $(HOMEDIR)/output

# 可以修改为其他的名字
APPNAME = "socketserver"

# 初始化命令变量
# 开发环境下不需要bcloud下载，直接指定开发环境下GOROOT路径，并注释BCLOUD下载依赖
GOROOT  := $(HOMEDIR)/../../../go/go1.13.5
GO      := $(GOROOT)/bin/go
GOPATH  := $(shell $(GO) env GOPATH)
GOMOD   := $(GO) mod
GOBUILD := $(GO) build
GOTEST  := $(GO) test
GOPKGS  := $$($(GO) list ./...| grep -vE "vendor")
# 执行编译，可使用命令 make 或 make all 执行, 顺序执行prepare -> compile -> test -> package 几个阶段
all: prepare compile package
# prepare阶段, 使用bcloud下载非Go依赖，使用GOD下载Go依赖, 可单独执行命令: make prepare
prepare: prepare-dep
prepare-dep:
    # 设置git， 保证github mirror能够下载
	git config --global http.sslVerify false
set-env:
	$(GO) env -w GONOPROXY=\*\*.baidu.com\*\*
	$(GO) env -w GOPROXY=http://goproxy.baidu-int.com
	$(GO) env -w GONOSUMDB=\*
#complile阶段，执行编译命令,可单独执行命令: make compile
compile:build
	$(GO) build
build: set-env
    # 下载Go依赖
	$(GOMOD)  download
# 	$(GOBUILD) agile构建环境无法满足要求，修改成本地docker编译后，打包到代码库，docker镜像请看README
#test阶段，进行单元测试， 可单独执行命令: make test
test: test-case
test-case: set-env
	$(GOTEST) -v -cover $(GOPKGS)
#与覆盖率平台打通，输出测试结果到文件中
#@$(GOTEST) -v -json -coverprofile=coverage.out $(GOPKGS) > testlog.out
#package阶段，对编译产出进行打包，输出到output目录, 可单独执行命令: make package
package: package-bin
package-bin:
#   灵力视频中台上线打包方式
# 	rm -rf $(OUTDIR)
# 	mkdir -p $(OUTDIR)/bin
# 	mkdir -p $(OUTDIR)/$(APPNAME)
# 	if [ -d "bin"  ]; then cp -r bin $(OUTDIR); fi
# 	mv $(APPNAME) $(OUTDIR)/bin
# 	if [ -d "conf"  ]; then cp -r conf $(OUTDIR)/bin/conf; fi
# 	if [ -d "data"  ]; then cp -r data $(OUTDIR)/bin/data; fi
# 	if [ -d "libs"  ]; then cp -r libs $(OUTDIR)/bin/libs; fi
# 	if [ -d "lib"   ]; then cp -r lib $(OUTDIR)/bin/lib; fi
# 	if [ -d "test"  ]; then cp -r test $(OUTDIR)/test_tools; fi
# 	if [ -f "control.sh" ]; then cp -r control.sh $(OUTDIR)/bin; fi
# 	if [ -d "script"  ]; then cp -r script/supervise $(OUTDIR)/bin; fi
# 	cd $(OUTDIR) && mv `ls | grep -v $(APPNAME)` $(APPNAME) && tar -czvf $(APPNAME).tar.gz ./* && `ls | grep -v $(APPNAME).tar.gz | xargs rm -rf`

#   opera打包方式
	rm -rf $(OUTDIR)
	mkdir -p $(OUTDIR)/bin
	if [ -d "bin"  ]; then cp -r bin $(OUTDIR); fi
	mv $(APPNAME) $(OUTDIR)/bin
	if [ -d "conf"  ]; then cp -r conf $(OUTDIR)/bin/conf; fi
	if [ -d "data"  ]; then cp -r data $(OUTDIR)/bin/data; fi
	if [ -d "libs"  ]; then cp -r libs $(OUTDIR)/bin/libs; fi
	if [ -d "lib"   ]; then cp -r lib $(OUTDIR)/lib; fi
	if [ -f "control.sh" ]; then cp -r control.sh $(OUTDIR); fi
	if [ -d "script"  ]; then cp -r script/* $(OUTDIR); fi
	cd $(OUTDIR) && tar -czvf $(APPNAME).tar.gz ./* && `ls | grep -v $(APPNAME).tar.gz | xargs rm -rf`

#clean阶段，清除过程中的输出, 可单独执行命令: make clean
clean:
	rm -rf $(OUTDIR)
# avoid filename conflict and speed up build
.PHONY: all prepare compile test package  clean build
