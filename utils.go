package detector

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func getWorkDir() string {
	var err error
	// 可执行文件所在目录
	appPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err == nil {
		if dirExist(appPath) {
			return appPath
		}
	}

	// 执行目录（兼容go run）
	workPath, err := os.Getwd()
	if err != nil {
		if dirExist(workPath) {
			return workPath
		}
	}
	return ""
}

func fileExist(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func dirExist(name string) bool {
	if fi, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	} else if !fi.IsDir() {
		return false
	}
	return true
}

func fileJoin(elm ...string) string {
	if p := filepath.Join(elm...); fileExist(p) {
		return p
	}
	return ""
}

func dirJoin(elm ...string) string {
	if p := filepath.Join(elm...); dirExist(p) {
		return p
	}
	return ""
}

func dirJoinDbg(elm ...string) string {
	if p := filepath.Join(elm...); dirExist(p) {
		return p
	}
	return ""
}

// 支持更复杂的函数名拆分
func nameSplit(s string) []string {
	res := make([]string, 0, 10)
	elmList := strings.Split(s, "_")
	var lastLink int
	for _, elm := range elmList {
		elmRunes := []rune(elm)
		resElmRunes := make([]rune, 0, 10)
		for _, r := range elmRunes {
			switch true {
			case isUpper(r):
				switch true {
				case len(resElmRunes) == 0:
					// 继续拼接
					resElmRunes = append(resElmRunes, r)
				case len(resElmRunes) > 0:
					lastElmRune := resElmRunes[len(resElmRunes)-1]
					if isUpper(lastElmRune) {
						// 上一个字母也是大写，继续拼接
						resElmRunes = append(resElmRunes, r)
					} else {
						// 上个字母不是大写，分割并重置
						if !isUpper(resElmRunes[0]) && len(res) > 0 {
							// 如果不是大写则拼接连字符
							res[len(res)-1] = res[len(res)-1] + strings.Repeat("_", lastLink+1) + string(resElmRunes)
						} else {
							res = append(res, string(resElmRunes))
						}

						resElmRunes = resElmRunes[:1]
						resElmRunes[0] = r
					}
				}
			default:
				// 继续拼接
				if l2 := len(resElmRunes) - 2; l2 >= 0 {
					lastElmRune := resElmRunes[l2+1]
					last2ElmRune := resElmRunes[l2]
					if isUpper(lastElmRune) && isUpper(last2ElmRune) {
						// 上两个都是大写，则切割
						res = append(res, string(resElmRunes[:l2+1]))
						resElmRunes = resElmRunes[l2+1:]
					}
				}
				resElmRunes = append(resElmRunes, r)
			}
		}

		// 收尾
		if len(resElmRunes) > 0 {
			if !isUpper(resElmRunes[0]) && len(res) > 0 {
				// 如果不是大写则拼接连字符且不是开头第一个单词
				res[len(res)-1] = res[len(res)-1] + strings.Repeat("_", lastLink+1) + string(resElmRunes)
			} else {
				res = append(res, string(resElmRunes))
			}
			lastLink = 0
			resElmRunes = resElmRunes[:0]
		} else {
			// 额外连字符
			lastLink++
		}
	}

	return res
}

func nameSplitToVar(s string) string {
	splits := nameSplit(s)
	if len(splits) == 0 {
		return ""
	}
	splits[0] = strings.ToLower(splits[0])
	return strings.Join(splits, "")
}

func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

var (
	envReplaceKeyReg    = regexp.MustCompile(`([^\w])`)
	envPfxReplaceKeyReg = regexp.MustCompile(`(^\d+)`)
)

func toEnvKey(s string) string {
	// 因为环境变量键只支持字母、数字、下划线，所以把字符串中的非法字符都处理成_
	s = envReplaceKeyReg.ReplaceAllString(s, "_")
	// 再把开头的数字处理掉（如果有）
	return envPfxReplaceKeyReg.ReplaceAllString(s, "_")
}

var envKeyReg = regexp.MustCompile(`^[_a-zA-Z][\w]+$`)

func isValidateEnvKey(s string) bool {
	// 因为环境变量键只支持字母、数字、下划线，且数字不能作为开头
	return envKeyReg.MatchString(s)
}
