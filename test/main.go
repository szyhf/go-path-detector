package main

import (
	"encoding/json"
	"log"

	"github.com/szyhf/go-path-detector"
)

type Work struct {
	// 测试普通的目录查找
	Path string
	// 测试自定义环境变量
	Conf Conf `pd:"Key(CONF_DIR);"`
	// 测试默认推导
	Runtimes Runtimes
	// 测试推断
	InferFile string `pd:"Infer()"`
}

type Conf struct {
	// 测试没有设置Path时自动跳过
	// Path string
	// 测试自定义的分隔符及后缀
	DitFile string `pd:"Ext(txt);Split(-);"`
	// 测试自定义的环境变量注入
	DBConfigID string `pd:"Key(DB_CNF_ID);"`
	// 测试带优先级的文件搜索
	LogID string `pd:"Ext(file);Priority(/root/go/src/github.com/szyhf/go-path-detector/test/priority_test);"`
}

type Runtimes struct {
	Path string
	Log  Log
	// 测试自定义的路径变量
	Search struct {
		DIYPath  string
		Document struct {
			Path string
			// 测试可以不存在的文件
			OptFile string `pd:"Opt();"`
		}
		Post struct {
			Path string
		}
	} `pd:"Path(DIYPath)"`
}

type Log struct {
	Path string
}

func main() {
	// log.SetFlags(// log.Ldate | // log.Ltime | // log.Llongfile)
	// detector.SetLogger(os.Stdout, "[DEBUG]", log.Ldate|log.Ltime|log.Llongfile)

	// 嵌套组合可以更快捷的优化组装
	work := &struct {
		Work
	}{}

	detector := detector.NewDetector().
		WithEnvPrefix("ENV").
		WithDirEnvKey("ENV_DIR") //.Debug(os.Stdout)
	// WithFileSplit("-")
	err := detector.
		Detect(&work.Work)
	if err != nil {
		data, _ := json.MarshalIndent(work, "", "\t")
		log.Printf("%+v", string(data))
		log.Fatal(err)
	}
	data, _ := json.MarshalIndent(work, "", "\t")
	log.Printf("%s", string(data))
}
