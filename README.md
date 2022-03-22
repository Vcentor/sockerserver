# 项目名称
虚拟数字人服务

## 快速开始
如何构建、安装、运行
一、vta算法构建：
```
1、拉取镜像
docker pull iregistry.baidu-int.com/art-cloud/centos7-gcc82-opera:vta-v2-rel-iv1.0.1
2、启动实例
docker run -it -v 本地算法路径:docker中映射路径 imageName /bin/bash
3、在docker映射路径编译算法
```

二、digital-human-server 构建
```
1、说明：
由于agile编译机的gcc8.2无法编译算法，所以采用docker中编译好可执行文件发不到agile上
2、编译方法：
docker pull iregistry.baidu-int.com/art-cloud/centos7-digitalhuman-golang-compile:1.0.0
docker run -it -v 本地go代码路径:docker中映射路径 -p 映射出的端口:docker中业务端口 imageName /bin/bash 
```

三、安装
使用make -f Makefile即可

四、注意
启动程序时，digitial-human-server/libs/vta/work_dir/lib中的库是vta算法自动加载的，
所以放置的目录需要和程序启动目录在同一级

## 测试
如何执行自动化测试

## 如何贡献
贡献patch流程、质量要求

## 讨论
百度Golang交流群：1450752

## 链接
[百度golang代码库组织和引用指南](http://wiki.baidu.com/pages/viewpage.action?pageId=515622823)
[百度内Go Module使用指南](http://wiki.baidu.com/pages/viewpage.action?pageId=917601678)

