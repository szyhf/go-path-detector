package detector

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
)

type Detector interface {
	// 根据传入的结构体进行搜索
	Detect(i interface{}) error
	// 统一设置所有环境变量的前缀
	WithEnvPrefix(prefix string) Detector
	// 统一设置所有文件的分隔符，默认为`.`
	WithFileSplit(split string) Detector
	// 会尝试优先根据`dirEnv`的配置值设置工作目录
	WithDirEnvKey(dirEnv string) Detector

	// 在搜索的同时打印搜索逻辑到log
	Debug(w io.Writer) Detector
}

func NewDetector() Detector {
	return &detector{
		envPrefix: "",

		dirEnvKey:              "",
		directoryNameParseType: NameParseType.SmartSnake,

		fileSplit:         ".",
		fileNameParseType: NameParseType.SmartSnake,
	}
}

type NameParseTypeID = int8

var NameParseType = struct {
	FieldName NameParseTypeID
	// 智能蛇形
	SmartSnake NameParseTypeID
}{
	FieldName:  1,
	SmartSnake: 2,
}

type detector struct {
	isDebug bool

	envPrefix string

	directoryNameParseType NameParseTypeID
	dirEnvKey              string
	dirSplit               string
	dirPrioirity           []string

	fileNameParseType NameParseTypeID
	fileSplit         string
	// baseDir string
}

func (this *detector) Detect(i interface{}) error {
	var err error
	t := reflect.TypeOf(i)
	v := reflect.ValueOf(i)
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("%T不是Ptr", i)
	}

	// 首先如果环境变量设置了，则只使用环境变量，失败了就报错
	if baseDir := this.getBaseDirByEnv(); baseDir != "" {
		// log.Println("Env BaseDir", baseDir)
		if err = this.tryDetector(baseDir, v); err == nil {
			return nil
		} else {
			return fmt.Errorf("无法根据根据环境变量%s提供的参数%s推导: %s", this.dirEnvKey, baseDir, err.Error())
		}
	}

	// 先用os.Args[0]尝试（可执行文件所在的目录尝试）
	if baseDir := this.getBaseDirByOSArgs(); baseDir != "" {
		// log.Println("ArgsBaseDir", baseDir)
		if err = this.tryDetector(baseDir, v); err == nil {
			return nil
		} else {
			err = fmt.Errorf("无法根据OSArgs[0]=%s推导：%s", baseDir, err.Error())
		}
	}
	// 再用命令执行目录尝试（兼容go run）
	if baseDir := this.getBaseDirByWD(); baseDir != "" {
		// log.Println("OSWDBaseDir", baseDir)
		if err2 := this.tryDetector(baseDir, v); err2 == nil {
			return nil
		} else {
			if err != nil {
				err = fmt.Errorf("%s;无法根据os.Getwd()=%s推导：%s", err.Error(), baseDir, err2.Error())
			} else {
				err = err2
			}
		}
	}

	return fmt.Errorf("无法找到工作目录，可能因为：%s", err.Error())
}

func (this *detector) tryDetector(baseDir string, v reflect.Value) error {
	dirSch, err := this.newDirSchema(v.Elem(), nil, nil)
	if err != nil {
		return err
	}
	dirSch.Path = baseDir
	err = dirSch.detector()
	if err != nil {
		return err
	}
	return nil
}

func (this *detector) WithEnvPrefix(prefix string) Detector {
	this.envPrefix = prefix
	return this
}

// 统一设置所有文件名的分隔符，默认为`.`
func (this *detector) WithFileSplit(split string) Detector {
	this.fileSplit = split
	return this
}

// 会尝试优先根据`dirEnv`的配置值设置工作目录
func (this *detector) WithDirEnvKey(dirEnv string) Detector {
	this.dirEnvKey = dirEnv
	return this
}

// 在搜索的同时打印搜索逻辑到logger
func (this *detector) Debug(output io.Writer) Detector {
	this.isDebug = true
	SetLogger(output, "[DEBUG]", 0)
	return this
}

func (this *detector) getBaseDirByEnv() string {
	// 找到工作目录
	if this.dirEnvKey != "" {
		// 通过环境变量获取
		if path := os.Getenv(this.dirEnvKey); path != "" {
			// log.Printf("os.Getenv(%s) = %s", path, this.dirEnvKey)
			return path
		}
	}
	return ""
}

// 可执行文件所在目录
func (this *detector) getBaseDirByOSArgs() string {
	appPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err == nil {
		if dirExist(appPath) {
			// log.Printf("os.Args[0] = %s", appPath)
			return appPath
		}
	}
	return ""
}

// 执行目录（兼容go run）
func (this *detector) getBaseDirByWD() string {
	workPath, err := os.Getwd()
	if err == nil {
		if dirExist(workPath) {
			// log.Printf("os.Getwd() = %s", workPath)
			return workPath
		}
	}
	return ""
}
