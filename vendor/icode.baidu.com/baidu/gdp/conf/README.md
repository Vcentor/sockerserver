# Conf 配置解析

## 快速开始
提供一个通用的，可扩展的的配置读取模块。  
如可读取常见的 .toml、.json 后缀的配置文件。  
并且可以扩展添加新的文件类型。


### 功能接口
```
// ParserFunc 对应文件后缀的配置解析方法
type ParserFunc func(bf []byte, obj interface{}) error
 
// BeforeFunc 辅助方法，在执行解析前，会先会配置的内容进行解析处理
type BeforeFunc func(conf Conf, content []byte) ([]byte, error)

// 读取并解析配置文件
Parse(confName string, obj interface{}) error

// 解析bytes内容
ParseBytes(fileExt string, content []byte, obj interface{}) error

// 配置文件是否存在
Exists(confName string) bool

// 注册一个指定后缀的配置的parser
// 如要添加 .ini 文件的支持，可在此注册对应的解析函数即可
RegisterParserFunc(fileExt string, fn ParserFn) error

// 注册一个辅助方法
RegisterBeforeFunc(name string,fn BeforeFunc) error
```



### Before回调
#### 1.系统环境变量替换
```
http_listen = "0.0.0.0:{env.LISTEN_PORT|8080}"
```
若环境变量中存在 `LISTEN_PORT=80`，最终得到的配置为：
```
http_listen = "0.0.0.0:80"
```
#### 2.应用环境信息替换
```
log="{gdp.LogDir}/mysql.log"
```

## 测试
```
go test --race ./...
```
## 如何贡献
## 讨论
上帝群号：1612141  