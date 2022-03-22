# Env-应用运行环境信息
   该模块定义应用所需要的一些基础信息，如 
`应用名称（AppName）`、
`当前机房(IDC)`、
`运行模式（RunMode）`、
`应用根目录(RootDir)`、
`配置文件目录(ConfDir)`、
`数据根目录(DataDir)`。  
   应用在启动阶段，首先将这些公共的基础信息进行初始化，其他功能模块则可以直接使用，以达到解耦合的目的。
   
## Getting Started|快速开始

### 1.1 import
```
import(
    "icode.baidu.com/baidu/gdp/env"
)
```

### 1.2 提供接口
#### 1.2.1 env 包的选项和方法

```
// Option 可配置选项
type Option struct {
	AppName string
	IDC     string
	RunMode string

	RootDir     string
	DataDir string
	LogDir  string
	ConfDir string
}
```

如下方法可直接使用(如 `env.RootDir()`)：
```

// 获取应用的根目录
RootDir() string


// 获取 应用所在idc
IDC() string


// 获取应用运行模式
RunMode() string


// 获取应用名称
AppName() string


// 获取配置文件更目录路径，默认: RootPath()/conf
ConfDir() string

// 获取数据文件根目录路径，默认: RootPath()/data
DataDir() string


// 获取日志根目录，默认：RootPath()/log
LogDir() string

// 获取当前环境的选项
Options() Option

// 复制一个新的env对象，并将传入的Option merge进去
CloneWithOption(opt Option) AppEnv
```
注意：
 * 上述方法为默认全局环境变量(Default)。
 * 若应用有特殊需求，可创建自己的环境变量。

其他系统信息：
```
// 本机ip，若获取失败，值为unknown
LocalIP() string 

PIDString() string
```
 
#### 1.2.2 多套环境信息支持
提供如下方法：
```
// 创建一套新的环境变量
New(opt Option) AppEnv
```

对于基础模块建议创建自己的环境信息实例，而不要直接使用全局的,若无设置再使用默认全局的。  

### 1.3 自动推断路径
  但是经常我们需要在子包中(如service/data/home/info)运行一个单测,而在此时，我们子包中的环境变量信息是不完整的。  
  故这时候若有依赖特定的配置文件(读配置文件),或者是写日志，将导致配置文件读取不到（配置文件的路径找不到），或者是日志写的路径不对。  
  当然也可以采用mock的方式来运行单测，但是写case的成本增加很多。  

目前env的`RootDir`,`DataDir`,`ConfDir`,`LogDir` 的自动推动策略如下：

>  1. 当前或者上级目录(向上级目录循环查找)，存在go.mod 或者 conf/app.toml,则其所在目录为RootDir。
>  2. 若上述都没有值，则RootDir为当前目录。
   
## Running the tests|测试
```
go test -cover ./...
```

## Contributing|如何贡献
> 1. 在本项目空间创建issue
> 2. 拉取分支，提交代码 （单测覆盖率达到100%）
> 3. cr 并merge代码


推荐代码Format工具: https://github.com/fsgo/go_fmt

## Discussion|讨论
上帝群号：1612141  
作者：duwei04