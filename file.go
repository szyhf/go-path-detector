package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

type fileSchema struct {
	_detector *detector

	fieldTag envTag
	field    reflect.StructField
	val      reflect.Value

	// 期望的文件名
	Name string
	// 文件的实际路径
	Path string
	// 文件所属的目录
	ParentDir *dirSchema `json:"-"`

	// 当前文件路径对应的环境变量名
	EnvPathKey string
}

func (this *fileSchema) detector() (err error) {
	// 如果没有预先设置的Path
	if this.Path == "" {
		this.Path, err = func() (string, error) {
			// 1. 根据当前目录对应的环境变量名
			if path := os.Getenv(this.EnvPathKey); path != "" {
				if fileExist(path) {
					return path, nil
				} else {
					// 如果配置了环境变量，则在出错时立即返回
					return "", fmt.Errorf("环境变量'%s'='%s'对应的文件不存在", this.EnvPathKey, path)
				}
			}

			// 2. 根据优先级目录
			for _, path := range this.fieldTag.Priority {
				if curPath := fileJoin(path, this.Name); curPath != "" {
					return curPath, nil
				}
			}
			// 3. 根据父目录
			if this.ParentDir != nil {
				if curPath := fileJoin(this.ParentDir.Path, this.Name); curPath != "" {
					return curPath, nil
				}
			}
			return "", fmt.Errorf("找不到%s的实际路径", this.Name)
		}()
		if err != nil {
			return err
		}
	}

	this.val.SetString(this.Path)

	return nil
}

func (this *fileSchema) initName() {
	if this.Name == "" {
		// 1. 从tag的Name字段获取
		if this.fieldTag.Name != "" {
			this.Name = this.fieldTag.Name
			return
		}
		// 2. 从字段名获取
		switch this._detector.fileNameParseType {
		case NameParseType.SmartSnake:
			sl := nameSplit(this.field.Name)
			if len(sl) > 0 {
				split := this._detector.fileSplit
				if this.fieldTag.Split != "" {
					split = this.fieldTag.Split
				}
				this.Name = strings.Join(sl, split)
				this.Name = strings.ToLower(this.Name)
				if this.fieldTag.Ext != "" {
					// 扩展名总是用.做分隔符
					this.Name = this.Name + "." + this.fieldTag.Ext
				}
				return
			}
			fallthrough
		default:
			this.Name = this.field.Name
			return
		}
	}
}

func (this *fileSchema) initEnvKey() {
	if this.EnvPathKey == "" {
		// 1. 从tag的Key字段获取
		if this.fieldTag.Key != "" {
			this.EnvPathKey = this.fieldTag.Key
			return
		}

		// 2. 根据规则`${DIR_PREFIX}_${DIR_1}(_${DIR_2}_${FILE_NAME}_${FILE_EXT})`
		elmList := make([]string, 0, 10)
		// 把目录名先拼好，文件名最后弄
		curDir := this.ParentDir
		for curDir != nil {
			if curDir.Name != "" {
				elmList = append(elmList, curDir.Name)
			}
			curDir = curDir.ParentDir
		}
		if this._detector.envPrefix != "" {
			elmList = append(elmList, this._detector.envPrefix)
		}
		elmSort := sort.StringSlice(elmList)
		for i := 0; i < elmSort.Len()/2; i++ {
			elmSort.Swap(i, elmSort.Len()-i-1)
		}

		// 给文件名前边多一个下划线
		elmList = append(elmList, "", this.Name)
		this.EnvPathKey = strings.Join(elmList, "_")
		// // log.Printf("%+v", elmList)
		this.EnvPathKey = strings.ToUpper(this.EnvPathKey)
		// 因为环境变量只能使用数字字母下划线，所以统一处理掉不合法的命名。
		this.EnvPathKey = toEnvKey(this.EnvPathKey)
		// // log.Println(convert.MustJsonPrettyString(elmList), this.EnvPathKey)
	}
}

func (this *fileSchema) genDoc() string {
	s := "路径搜索逻辑："
	h1Idx := 0
	padStr := strings.Repeat(" ", len(this.fieldTag.Priority))
	if this.EnvPathKey != "" {
		h1Idx++
		s += fmt.Sprintf("\n%d、 %s从环境变量'%s'获取，如果文件存在立即返回。", h1Idx, padStr, this.EnvPathKey)
	}
	if len(this.fieldTag.Priority) > 0 {
		h1Idx++
		for i, priorityPath := range this.fieldTag.Priority {
			s += fmt.Sprintf("\n%d.%d、从优先路径'%s'获取，如果文件存在则立即返回。", h1Idx, i+1, filepath.Join(priorityPath, this.Name))
		}
	}

	if this.ParentDir != nil {
		h1Idx++
		s += fmt.Sprintf("\n%d、 %s从父目录'%s'获取，如果文件存在则立即返回。", h1Idx, padStr, filepath.Join(this.ParentDir.Path, this.Name))
	}

	return s
}

func (this *detector) newFileSchema(v reflect.Value, parentDir *dirSchema, f *reflect.StructField) (*fileSchema, error) {
	if v.Kind() != reflect.String {
		return nil, fmt.Errorf("指向文件路径仅可使用string类型")
	}
	fs := &fileSchema{
		_detector: this,
		val:       v,
		ParentDir: parentDir,
	}
	if f != nil {
		fs.field = *f
		fs.fieldTag = parseTag(f.Tag)
	} else {
		fs.fieldTag = defaultFileTag
	}
	fs.initName()
	fs.initEnvKey()
	return fs, nil
}
