package detector

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var (
	// Priority(/run/configs)
	tagReg = regexp.MustCompile(`^(\w+)\((\S*)\)$`)
)

type envTag struct {
	// 设置的文件名/目录名，优先级最高。
	// 如果设置了该Tag，则`Key`、`Ext`都会无效。
	Name string
	// 从环境变量获取Name的强制配置。
	// 如果设置了该项，且没有设置Name，
	// 则优先通过该变量尝试获取文件/目录名。
	// 如果获取失败还是会进行推断。
	Key string
	// 生成蛇形名称时的分隔符，默认为"_"
	Split string
	// 后缀，如果设置了Ext，则自动拼接文件名时总是在后缀前使用"."作为拼接符，而不是FileSplit。
	// 以便自动生成形如"hello_world.txt"的格式。
	Ext string
	// 如果设置了该项，则该文件/目录可以不存在
	Opt bool
	// 仅对struct有效，如果设置了该项。
	// 则会把当前目录的路径写入该结构体的该名称的成员变量。
	// 该成员必须是string类型
	Path string
	// 优先搜索目录
	Priority []string
}

func parseTag(tag reflect.StructTag) envTag {
	s := tag.Get("pd")
	if len(s) == 0 {
		return envTag{
			Path:     "Path",
			Priority: []string{},
		}
	}
	et := envTag{}
	sl := strings.Split(s, ";")
	for _, s := range sl {
		if s == "" {
			continue
		}
		match := tagReg.FindStringSubmatch(s)
		switch match[1] {
		case "Name":
			et.Name = match[2]
		case "Key":
			et.Key = match[2]
			switch true {
			case et.Key == "-":
				// 要求不使用环境变量
			case et.Key != "":
				if !isValidateEnvKey(et.Key) {
					panic(fmt.Sprintf("非法的环境变量名'%s'，必须由字母、数字或下划线组成，且数字不是开头。", et.Key))
				}
			}
		case "Split":
			et.Split = match[2]
		case "Ext":
			et.Ext = match[2]
		case "Opt":
			et.Opt = true
		case "Priority":
			et.Priority = strings.Split(match[2], "|")
		case "Path":
			et.Path = match[2]
		default:
			panic("未知的tag：" + match[1])
		}
	}

	if et.Path == "" {
		et.Path = "Path"
	}
	return et
}

var defaultDirectoryTag = envTag{
	Path:     "Path",
	Priority: []string{},
}

var defaultFileTag = envTag{
	Path:     "Path",
	Priority: []string{},
}
