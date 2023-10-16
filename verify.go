// 验证器，封装实现gin框架的验证器功能
package verify

import (
	"errors"
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
	//"sync"
)

// 验证数据
// param：传指针
// args：第一个参数是校验场景，第二个参数是语言
func Check(param interface{}, args ...string) (pass bool, warnInfo map[string]string, err error) {
	var (
		scene string // 验证场景
		lang  = "zh" // 语言,默认中文
	)
	if len(args) > 0 {
		scene = args[0]
	}
	if len(args) > 1 && args[1] != "" {
		lang = args[1]
	}
	paramType := reflect.TypeOf(param)
	paramValue := reflect.ValueOf(param)
	validate, ok := binding.Validator.Engine().(*validator.Validate)

	if !ok {
		return false, nil, errors.New("verify failed")
	}

	// 注册一个获取field别名的自定义方法，可将field名替换成别名，
	// 生成的告警信息中，key和msg里的field名都回被替换，这里将json的值作为别名
	// 同时记录对应的byname
	byNames := make(map[string]string)
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return ""
		}

		byName := fld.Tag.Get("byname")
		if byName != "" {
			byNames[name] = byName
		}
		return name
	})

	zhT := zh.New() //中文翻译器
	enT := en.New() //英文翻译器

	// 第一个参数是备用（fallback）的语言环境
	// 后面的参数是应该支持的语言环境（支持多个）
	// uni := ut.New(zhT, zhT) 也是可以的
	uni := ut.New(enT, zhT, enT)

	// 也可以使用 uni.FindTranslator(...) 传入多个locale进行查找
	trans, _ := uni.GetTranslator(lang)
	// 注册翻译器
	switch lang {
	case "en":
		err = enTranslations.RegisterDefaultTranslations(validate, trans)
	case "zh":
		fallthrough
	default:
		err = zhTranslations.RegisterDefaultTranslations(validate, trans)
	}
	if err != nil {
		return
	}

	// 设置用于校验的tag name
	//validate.SetTagName("verify")

	// 在校验器注册自定义的校验方法
	for i := 0; i < paramType.NumMethod(); i++ {
		method := paramType.Method(i)
		if strIn(method.Name, []string{"Scenes", "Warns"}) {
			// 去掉不是自定义验证的方法
			//Warns()  //获取错误的返回信息
			//Scenes() //所有校验场景
			continue
		}
		verifyFunc := paramValue.MethodByName(method.Name).Interface()
		switch vef := verifyFunc.(type) {
		case func(reflect.Value) bool:
			validate.RegisterValidation(method.Name, func(vf validator.FieldLevel) bool {
				return vef(vf.Field())
			})
		case func(validator.FieldLevel) bool:
			validate.RegisterValidation(method.Name, vef)
		}
		//if err := v.RegisterValidation(methodName, methodFunc); err != nil {
		//	return err
		//}
	}

	//自定义翻译提示
	if warnsFunc := paramValue.MethodByName("Warns"); warnsFunc.IsValid() {
		if warns, ok := warnsFunc.Call(nil)[0].Interface().(map[string]string); ok {
			for key, warn := range warns {
				//闭包，避免使用环境变量时key和val的值变了
				func(key string, msg string) {
					_ = validate.RegisterTranslation(
						key,
						trans,
						func(ut ut.Translator) error {
							return ut.Add(key, msg, true)
						},
						func(ut ut.Translator, fe validator.FieldError) string {
							t, _ := ut.T(key, fe.Field())
							return t
						})
				}(key, warn)
			}
		} else {
			err = errors.New("Warns() returned parameter format error")
			return
		}
	}

	// 获取需要验证的字段，空的话就是全部
	isVerifyPartial, verifyFields, err := getVerifyFields(scene, paramValue)
	if err != nil {
		return
	}

	// 开始校验
	var errs error
	if isVerifyPartial {
		// 根据场景，按字段选择性校验
		var fields = make([]string, 0)
		switch vFields := verifyFields.(type) {
		case []string:
			fields = vFields
		case map[string]string:
			for k, _ := range vFields {
				fields = append(fields, k)
			}
		}
		if len(fields) > 0 {
			errs = validate.StructPartial(param, fields...)
		}
	} else {
		errs = validate.Struct(param)
	}
	warnInfo = make(map[string]string)
	if errs != nil {
		validateErrs, _ := errs.(validator.ValidationErrors)

		for _, validateErr := range validateErrs {

			namespace := validateErr.Namespace()
			key := namespace[strings.Index(namespace, ".")+1:]

			errMsg := validateErr.Translate(trans)
			// 上面将RegisterTagNameFunc()方法将一些关键词替换成field别名后，进一步将警告信息里的别名转换成需要的名称byname（Tag）
			if bynames := getBynames(paramType.Elem(), validateErr.StructNamespace()); bynames != nil && len(bynames) > 0 {
				for name, byname := range bynames {
					reg := regexp.MustCompile(`(?i)` + name)
					errMsg = reg.ReplaceAllString(errMsg, byname)
				}
			}
			warnInfo[key] = errMsg
		}
	}

	// 整体性校验，一些上面的方法校验不出来的话，可以在这个方法里校验
	if wholeCheckFunc := paramValue.MethodByName("WholeCheck"); wholeCheckFunc.IsValid() {
		paramList := []reflect.Value{reflect.ValueOf(scene)}
		if warns, ok := wholeCheckFunc.Call(paramList)[0].Interface().(map[string]string); !ok {
			err = errors.New("WholeCheck() returned parameter format error")
		} else {
			for k, w := range warns {
				if _, exist := warnInfo[k]; !exist {
					warnInfo[k] = w
				}
				/*else {
					warnInfo[k] += ";"+w
				}*/
			}
		}
	}

	if len(warnInfo) > 0 {
		return false, warnInfo, nil
	}

	pass = true
	return
}

func strIn(target string, strArr []string) bool {
	for _, str := range strArr {
		if target == str {
			return true
		}
	}
	return false
}

func getVerifyFields(scene string, paramValue reflect.Value) (isVerifyPartial bool, verifyFields interface{}, err error) {
	if scenesFunc := paramValue.MethodByName("Scenes"); scenesFunc.IsValid() {
		if scenes, ok := scenesFunc.Call([]reflect.Value{})[0].Interface().(map[string]interface{}); !ok {
			err = errors.New("Scenes() returned parameter format error")
		} else if scene != "" && scenes[scene] != nil {
			verifyFields = scenes[scene]
			isVerifyPartial = true
		}
	}

	return
}

func getBynames(paramType reflect.Type, namespace string) map[string]string {
	bynames := make(map[string]string)

	path := strings.Split(namespace, ".")
	if len(path) < 2 {
		// 没拆分出多个的就是找不到的，返回空值
		return bynames
	}

	path = path[1:] // 去掉第一个，即最外层的strut名
	pType := paramType
	reg := regexp.MustCompile(`\[([0-9]+)\]$`)
	for index, step := range path {
		// 先判断是不是数组结构
		pos := reg.FindIndex([]byte(step))
		if len(pos) > 0 {
			step = step[:pos[0]]
		}
		field, ok := pType.FieldByName(step)

		if !ok {
			// 找不到变量直接返回
			return bynames
		}

		if index+1 == len(path) {
			// 到达需要的最底层的field
			name := field.Tag.Get("json")
			bynameStr := field.Tag.Get("byname")
			if bynameStr == "" {
				break
			}
			bynameArr := strings.Split(field.Tag.Get("byname"), ",") // 可能会有多个name的别名
			for _, byname := range bynameArr {
				val := strings.Split(byname, ":")
				if len(val) > 1 {
					bynames[val[0]] = val[1]
				} else {
					bynames[name] = val[0]
				}
			}
			break
		}

		// 继续深入
		pType = field.Type
		if len(pos) > 0 {
			// 数组要多加个Elem()
			pType = pType.Elem()
		}
	}
	return bynames
}
