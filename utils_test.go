package detector

import (
	"testing"
)

func TestToEnvKey(t *testing.T) {
	res := toEnvKey("Hello-World_Foo你好Bar.txt")
	if res != "Hello_World_Foo__Bar_txt" {
		t.Error("预料之外的结果：", res)
	}
}

func TestSplit(t *testing.T) {
	// 1. 小写->大写时分割
	// 1. 小写->小写持续
	// 1. 小写->连字符持续

	// 1. 连字符->连字符跳过
	// 1. 连字符->大写分割
	// 1. 连字符->小写持续

	// 1. 大写->大写持续
	// 1. 大写->小写，如果单大写则持续
	// 1. 大写->连字符分割
	exp := []string{
		"hello", "World", "Foo", "ID", "Third", "APP", "Bar", "You", "Hu_hu_yo__xi", "XI",
	}
	for i, act := range nameSplit("helloWorld_FooIDThirdAPP_Bar__You_Hu_hu_yo__xiXI") {
		if exp[i] != act {
			t.Errorf(`Exp[%d]"%s"==Act[%d]"%s"`, i, exp[i], i, act)
			t.Fail()
		}
	}

	if nameSplit("user")[0] != "user" {
		t.Errorf(`Exp[0]"user"==Act[0]"user"`)
		t.Fail()
	}
}
