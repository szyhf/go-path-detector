# 说明
[![Build Status](https://travis-ci.org/szyhf/go-path-detector.svg?branch=master)](https://travis-ci.org/szyhf/go-path-detector)
[![Go Report Card](https://goreportcard.com/badge/github.com/szyhf/go-path-detector)](https://goreportcard.com/report/github.com/szyhf/go-path-detector)

根据特定规则自动发现当前应用运行的目录环境（宿主主机/操作系统），找到相关目录、文件的真实路径。

## 例子

更多内容请参考test包及后续说明。

```golang
type Work struct {
	// 测试普通的目录查找
	Path string
	// 测试自定义环境变量
	Conf struct {
		// 测试没有设置Path时自动跳过
		// Path string
		// 测试自定义的分隔符及后缀
		DitFile string `pd:"Ext(txt);Split(-);"`
		// 测试自定义的环境变量注入
		DBConfigID string `pd:"Key(DB_CNF_ID);"`
		// 测试带优先级的文件搜索
		LogID string `pd:"Ext(file);Priority(/root/go/src/github.com/szyhf/go-path-detector/test/priority_test);"`
	} `pd:"Key(CONF_DIR);"`
}
func main(){
	work := &Work{}
	err := detector.NewDetector().Detect(work)
	if err == nil {
		// ...
	}
}
```

```json
{
	"Path": "/root/go/src/github.com/szyhf/go-path-detector/test",
	"Conf": {
		"DitFile": "/root/go/src/github.com/szyhf/go-path-detector/test/conf/dit-file.txt",
		"DBConfigID": "/root/go/src/github.com/szyhf/go-path-detector/test/priority_test/db.conf.test.hello.id",
		"LogID": "/root/go/src/github.com/szyhf/go-path-detector/test/priority_test/log.id.file"
	}
}
```

> 主要为了解决使用容器时与集群环境之间的映射问题。

可以按指定规则发现指定路径，自动生成环境变量名，自动判定环境中目录、文件是否存在，并提供一些控制选项。

```go
// 简单例子
type dir struct {
	Conf struct {
		// 搜索到的目录
		Path string

		DBYaml        string `pd:"Opt();"`
		GRPC      string `pd:"Ext(json);"`
		// 子目录
		CA struct{
			Path string
		}
	} `pd:"Priority(/run/configs)"`
}
func init() {
	var Dir dir
	det := detector.NewDetector().
		WithEnvPrefix("STH").
		// 可以通过该方法打印搜索逻辑
		// 用于调试
		Debug(os.Stdout)
	err := det.Detect(&Dir)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(Dir)
}
```

## 名称推断

名称推断指如何根据一个成员变量的名称推断该目录、文件的名字。

为了最简化操作，默认的名称推断为我自己实现的`SmartSnake`，基本规则是蛇形，但对连续因为缩写所以使用连续的大写字母支持更符合直觉，例如`UserID`=>`user_id`、`APPUser`=>`app_user`等，可以查看相关测试。

> 可以通过配置选择直接使用成员变量名，或者通过`tag`实现更丰富的配置（但对于项目来说，规则越通用，特例越少越好）。

## 环境变量

因为本库的核心目的是解决应用在部署环境运行时，可以根据更灵活\配置化的逻辑，找到运行环境的指定目录、文件。但考虑到实际部署环境可能存在特殊情况，不适合根据规则推断，所以搜索路径都可以通过`环境变量`直接注入，环境变量注入的值会看做部署人员对当前运行环境认可结果，如果基于此探测失败，且该目标不是可选目标，会立即报错并返回。

默认情况下，环境变量的生成规则如下：

+ 某个目录，如例子中的`CA`，绑定的环境变量为`CONF_CA`。
+ 某个文件，如例子中的`DBYaml`，绑定的环境变量为`CONF__DB_YAML`，即文件前会多一个`_`进行区分。

> 由于主流的sh中环境变量只支持字母、数字、下划线，且数字不能作为开头，所以这里只提供`tag`的方式强制指定环境变量。
> 另外可以通过执行`WithEnvPrefix(...)`的方法强制给所有自动生成的环境变量名增加固定前缀，会自动补`_`，如例子中的最终结果会是`STH_CONF_CA`等。

## 调试

因为规则比较复杂，可以通过调用`Debug(...)`的方法，将搜索流程打印出来（仅会打印到出错的地方为止），以便参考。

### 初始化工作目录

1. WithDirEnv("dir_env")设置的DirEnv环境变量，如`WPLAY_DIR`，如果配置了DirEnv，但从该变量取值指向的目录不存在，则立即报错。
1. 使用可执行文件所在目录(`os.Args[0]`)作为工作目录进行搜索尝试。
1. 使用执行命令时的目录(`os.Getwd()`)，主要用于兼容`go run`逻辑。

### 环境变量的自动规则

`${DIR_PREFIX}_${DIR_1}(_${DIR_2}_${FILE_NAME}_${FILE_EXT})`

例如，`WPLAY_CONFIG`表示`./config`目录的实际路径；
`WPLAY_CONFIG__DB_YAML`表示`./config/db.yaml`的实际路径。

> 默认环境变量生成结果会统一转成全大写。

### 子级目录推断优先级

1. 根据环境变量读入当前目录路径：如果读入的路径有效，则返回；如果读入的路径非法或者不存在，则继续。
1. 根据`tag`设置`Priority(...)`尝试组装子级目录路径，直到某个路径有效，或者全部失败为止。
1. 根据父级目录尝试组装子级目录，如果失败则报错。

通过`tag`可以配置相对路径优先级，默认返回当前

## Tag字段

### Path(field_name)

必须标识在struct类型的字段上，会将探索到的当前目录的路径写入到该struct的该名称的string类型成员上。

如果没有，会尝试写入字段名为`Path`且类型为`string`的成员变量。

```go
type Dir struct{
	ConfDir struct{
		Path string
	}
	SecretDir struct{
		Path string
		DirPath string
	} `pd:Path(DirPath);`
}
//（自动推导）ConfDir.Path = "path/to/conf_dir"
// SecretDir.DirPath = "path/to/secret_dir"
// SecretDir.Path会被看做该目录下应该有一个名为`Path`的文件。
```

### Name(target_name)

当前目录/文件的实际名称，如果设置了该变量，则其他配置都及自动推断的逻辑均无效。

```go
type Dir struct{
	ConfDir struct{
		DBFile string `pd:"Name(db.hello)"`
	} `pd:"Name(configs)"`
}
// 会尝试搜索名为`configs`的目录，如果找不到，则报错。
// 会尝试在`configs`目录下搜索名为`db.hello`的文件，如果找不到，则会报错。
```

### Key(env_path_key)

使用环境变量推断目录/文件的路径时，使用该项配置的值，并跳过自动生成环境变量名。

如果该变量的取值无效，会根据优先级继续按顺序执行后续推断。

> 如果不希望使用环境变量注入该路径，可以设置`-`表示跳过。

```go
type Dir struct{
	ConfDir struct{
		DBFile string `pd:"Key(Demo_DBFILE);"`
	}
}
// 会尝试通过名为`CONF_DIR`（自动推导）的环境变量获取ConfDir的路径。
// 会尝试通过名为`Demo_DBFILE`的环境变量获取DBFile的路径。
```

### Split(split_str)

根据蛇形规则生成目录/文件名时使用的分隔符。

> 默认目录名使用`_`，文件名使用`.`

```go
type Dir struct{
	ConfDir struct{
		DBConfig string `pd:"Split(-);"`
		IMConfig string
	}
	SecretDir struct{
	}`pd:"Split(-);"`
}
// 解析时会考虑文件名为`DB-Config`、`IM.Config`
// 考虑目录名为`conf_dir`、`secret-dir`
```

### Ext(file_ext)

因为文件后缀的分隔符总是'.'，如果Split不是'.'，使用Ext可以强制指定后缀，且该后缀总是用'.'跟前边的文件名进行拼接。

```go
type Dir struct{
	Conf struct{
		DBConfig string `pd:"Split(_);Ext(json);"`
	}
}
// 解析时会考虑文件名为`DB_Config.json`
```

### Priority(path_1|path_2)

使用`|`作为分隔符，以便传入多个路径，在搜索目录时，如果按照环境变量搜索失败，会逐个尝试根据该path来搜索目录，或者在目录下搜索文件，直到有一个成功或者全部失败为止。

```go
type Dir struct{
	Conf struct{
		DBConfig string `pd:"Priority(/run/conf|/run/config);"`
	} `pd:"Priority(/run/configs|/run/conf);"`
}
// 尝试环境变量获取失败后，会考虑Conf的目录为`/run/configs`，再考虑Conf的目录为`/run/conf`
// 尝试环境变量获取失败后，会优先考虑`/run/conf/DB.Config`，再考虑`/run/config`
```

### Opt()

无参数，设置了该选项后，则对应的目录或者文件考虑为可选项目，如果不存在不会报错。

> 因为搜索有多套推断逻辑，请不要让所有项目都是可选的，那么可能会推断为刚好目标项目都不存在！

### Infer()

无参数，设置了该选项后，如果当前目录/文件不存在，则会基于其父目录的路径和当前名称写入推断路径。

> 如果给目录设置了Infer，则目录如果不存在，但由于推断生成了路径时，会继续搜索子成员，可能会报错。