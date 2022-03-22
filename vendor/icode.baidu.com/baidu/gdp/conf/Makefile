# init project path
HOMEDIR := $(shell pwd)
OUTDIR  := $(HOMEDIR)/output

# init command params
GOROOT  := $(HOMEDIR)/../../../baidu/go-env/go1-14-linux-amd64/
GO      := $(GOROOT)/bin/go
GOPATH  := $(GO) env GOPATH
GOMOD   := $(GO) mod
GOBUILD := $(GO) build
GOTEST  := $(GO) test -gcflags="-N -l"
GOPKGS  := $$($(GO) list ./...| grep -vE "vendor")

# test cover files
COVPROF := $(HOMEDIR)/covprof.out  # coverage profile
COVFUNC := $(HOMEDIR)/covfunc.txt  # coverage profile information for each function
COVHTML := $(HOMEDIR)/covhtml.html # HTML representation of coverage profile

# make, make all
all: prepare compile package

# set proxy env
set-env:
	$(GO) env -w GO111MODULE=on
	$(GO) env -w GONOPROXY=\*.baidu.com\*
	$(GO) env -w GOPROXY=https://goproxy.baidu-int.com/
	$(GO) env -w GONOSUMDB=\*



# make prepare, download dependencies
prepare: bcloud gomod

bcloud:
	bcloud local -U
gomod: set-env
	$(GOMOD) tidy

# make compile, go build
compile: build

build: set-env
	$(GOBUILD)

# make test, test your code
test: test-case

test-case: 
	$(GOTEST) -cover -race -coverprofile=$(COVPROF) $(GOPKGS) 
	$(GO) tool cover -o $(COVFUNC) -func=$(COVPROF)
	$(GO) tool cover -o $(COVHTML) -html=$(COVPROF)

# make package
package: package-bin

package-bin:
	rm -rf $(OUTDIR)
	mkdir -p $(OUTDIR)

# make install
install: package

# make clean
clean:
	$(GO) clean
	rm -rf $(OUTDIR)
	rm -f  $(COVPROF) $(COVFUNC) $(COVHTML)

# avoid filename conflict and speed up build
.PHONY: all prepare compile test package install clean build 
