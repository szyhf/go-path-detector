package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

type directory struct {
	path string
}

type dirSchema struct {
	_detector *detector

	fieldTag envTag
	field    reflect.StructField
	fieldVal reflect.Value

	// 当前目录对应的环境变量名
	EnvPathKey string

	// 当前目录名
	Name string
	// 当前目录路径
	Path string
	// 对应结构体中用来存储Path的value
	pathFieldVal reflect.Value
	// 父目录
	ParentDir *dirSchema `json:"-"`
	// 子目录集合
	ChildrenDir []*dirSchema
	// 子文件集合
	ChildrenFile []*fileSchema
}

func (this *dirSchema) detector() (err error) {
	// 如果没有预先设置的Path
	if this.Path == "" {
		this.Path, err = func() (string, error) {
			// 1. 根据当前目录对应的环境变量名
			if this.EnvPathKey != "" {
				if path := os.Getenv(this.EnvPathKey); path != "" {
					if dirExist(path) {
						return path, nil
					} else {
						// 如果配置了环境变量，则在出错时立即返回
						return "", fmt.Errorf("环境变量'%s'='%s'对应的目录不存在", this.EnvPathKey, path)
					}
				}
			}
			// 2. 根据优先级目录
			for _, path := range this.fieldTag.Priority {
				// 目录则直接使用优先级目录作为目录尝试
				if dirExist(path) {
					return path, nil
				}
			}
			// 3. 根据父目录
			if this.ParentDir != nil {
				if curPath := dirJoin(this.ParentDir.Path, this.Name); curPath != "" {
					return curPath, nil
				}
			}
			return "", fmt.Errorf("找不到%s的实际路径", this.Name)
		}()
		if err != nil {
			return err
		}
	}

	if this.pathFieldVal.CanSet() {
		// 可能为空
		this.pathFieldVal.SetString(this.Path)
	}

	// 处理当前目录文件
	for i := 0; i < len(this.ChildrenFile); i++ {
		fileSch := this.ChildrenFile[i]
		err := fileSch.detector()
		if this._detector.isDebug {
			log.Println("\n" + fileSch.genDoc())
		}
		if err != nil {
			if !fileSch.fieldTag.Opt {
				return fmt.Errorf("处理%s下的文件%s出错{%s}", this.Name, fileSch.Name, err.Error())
			} else {
				// 该文件是可选的，则异常时移除
				this.ChildrenFile = append(this.ChildrenFile[:i], this.ChildrenFile[i+1:]...)
				// 因为移除了一个文件，所以要跳过本次自增
				i--
			}
		}
	}

	// 继续处理子目录
	for i := 0; i < len(this.ChildrenDir); i++ {
		dirSch := this.ChildrenDir[i]
		err := dirSch.detector()
		if this._detector.isDebug {
			log.Println("\n" + dirSch.genDoc())
		}
		if err != nil {
			if !dirSch.fieldTag.Opt {
				return fmt.Errorf("处理%s的子目录%s出错{%s}", this.Name, dirSch.Name, err.Error())
			} else {
				// 该文件是可选的，则异常时移除
				this.ChildrenDir = append(this.ChildrenDir[:i], this.ChildrenDir[i+1:]...)
				// 因为移除了一个目录，所以要跳过本次自增
				i--
			}
		}
	}
	return nil
}

func (this *dirSchema) genDoc() string {
	s := "目录搜索逻辑："
	h1Idx := 0
	padStr := strings.Repeat(" ", len(this.fieldTag.Priority))

	if this.EnvPathKey != "" {
		h1Idx++
		s += fmt.Sprintf("\n%d、 %s从环境变量'%s'获取，如果目录存在立即返回。", h1Idx, padStr, this.EnvPathKey)
	}

	if len(this.fieldTag.Priority) > 0 {
		h1Idx++
		for i, priorityPath := range this.fieldTag.Priority {
			s += fmt.Sprintf("\n%d.%d、从优先路径'%s'获取，如果目录存在则立即返回。", h1Idx, i+1, filepath.Join(priorityPath, this.Name))
		}
	}

	if this.ParentDir != nil {
		h1Idx++
		s += fmt.Sprintf("\n%d、 %s从父目录'%s'获取，如果目录存在则立即返回。", h1Idx, padStr, filepath.Join(this.ParentDir.Path, this.Name))
	}

	return s
}

func (this *detector) newDirSchema(v reflect.Value, parentDir *dirSchema, f *reflect.StructField) (*dirSchema, error) {
	t := v.Type()
	if v.Kind() == reflect.Ptr {
		return nil, fmt.Errorf("为了防止意外的循环引用，请不要使用Ptr类型")
	}

	curDirSch := &dirSchema{
		_detector:    this,
		ParentDir:    parentDir,
		fieldVal:     v,
		ChildrenDir:  make([]*dirSchema, 0, t.NumField()),
		ChildrenFile: make([]*fileSchema, 0, t.NumField()),
	}

	if f != nil {
		curDirSch.field = *f
		curDirSch.fieldTag = parseTag(f.Tag)
	} else {
		curDirSch.fieldTag = defaultDirectoryTag
	}
	curDirSch.initName()
	curDirSch.initEnvKey()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)
		if !fv.IsValid() {
			continue
		}
		switch f.Type.Kind() {
		case reflect.Struct:
			// 迭代
			childDirSch, err := this.newDirSchema(fv, curDirSch, &f)
			if err != nil {
				// log.Fatal(err)
			}
			curDirSch.ChildrenDir = append(curDirSch.ChildrenDir, childDirSch)
		case reflect.String:
			// 如果符合该目录标注的PATH字段
			if f.Name == curDirSch.fieldTag.Path {
				// 如果被标记为Path，则表示当前目录
				curDirSch.pathFieldVal = fv
			} else {
				// 如果没有被标记为Path，则表示当前目录下的某个文件
				fileSch, err := this.newFileSchema(fv, curDirSch, &f)
				if err != nil {
					// log.Fatal(err)
				}
				curDirSch.ChildrenFile = append(curDirSch.ChildrenFile, fileSch)
			}
		default:
			return nil, fmt.Errorf("不支持的数据类型%s", f.Type.Name())
		}
	}
	return curDirSch, nil
}

func (this *dirSchema) initName() {
	if this.Name == "" {
		// 1. 从tag的Name字段获取
		if this.fieldTag.Name != "" {
			this.Name = this.fieldTag.Name
			return
		}
		// 根目录另外处理
		if this.ParentDir == nil {
			// 根目录
			if this._detector.envPrefix != "" {
				this.Name = this._detector.envPrefix
			} else {
				// this.Name = filepath.Base(this._detector.getBaseDir())
			}
		}

		// 2. 从字段名获取
		switch this._detector.directoryNameParseType {
		case NameParseType.SmartSnake:
			sl := nameSplit(this.field.Name)
			if len(sl) > 0 {
				dirSplt := this._detector.dirSplit
				if this.fieldTag.Split != "" {
					dirSplt = this.fieldTag.Split
				}
				this.Name = strings.Join(sl, dirSplt)
				if this.fieldTag.Ext != "" {
					// 目录一般不要设Ext……
					this.Name = this.Name + "." + this.fieldTag.Ext
				}
				this.Name = strings.ToLower(this.Name)
				return
			}
			fallthrough
		default:
			this.Name = this.field.Name
			return
		}
	}
}

func (this *dirSchema) initEnvKey() {
	if this.EnvPathKey == "" {
		// 1. 从tag的Key字段获取
		if this.fieldTag.Key != "" {
			this.EnvPathKey = this.fieldTag.Key
			return
		}
		// 2. 根据规则`${DIR_PREFIX}_${DIR_1}(_${DIR_2}_${FILE_NAME}_${FILE_EXT})`
		elmList := make([]string, 0, 10)
		curDir := this
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
		this.EnvPathKey = strings.Join(elmList, "_")
		// // log.Println(convert.MustJsonPrettyString(elmList))
		this.EnvPathKey = strings.ToUpper(this.EnvPathKey)
		// // log.Println(this.EnvPathKey)
	}
}
