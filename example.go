package verify

import (
	"github.com/go-playground/validator/v10"
	"reflect"
)

// 需要校验的结构体的例子
// 关于更多的binding标签的校验规则：https://pkg.go.dev/github.com/go-playground/validator/v10
// byname标签：字段对应的别名，当输出告警信息中有json标签的值时，会被替换成别名
type Example struct {
	ID     int    `json:"id" binding:"required"`
	Title  string `json:"title" binding:"required" byname:"标题"`
	Author Author `json:"author"`
	Works  []Work `json:"works" binding:"gt=0,dive,required" byname:"公式"`
}

type Work struct {
	Name string `json:"name" byname:"名字"`
	Code string `json:"code" binding:"required" byname:"代码"`
}

type Author struct {
	FirstName string `json:"first_name" byname:"姓"`
	LastName  string `json:"last_name"  byname:"名"`
	Age       string `json:"age" binding:"CheckAge" byname:"年龄"`
}

// 当对应的规则没通过时，输出下面信息
// 有默认的信息，如果对默认信息不满意，可以用这个覆盖
func (t Example) Warns() map[string]string {
	return map[string]string{
		"required": "{0}是必须的",
		"CheckAge": "年龄不对",
	}
}

// 不同场景可以限定对不同的字段做校验
func (t Example) Scenes() map[string]interface{} {
	return map[string]interface{}{
		"create": []string{
			"Title", "Works",
		},
		"update": map[string]string{
			// 未实现，
			// 当是“”时，用结构体里的binging，不为空时，覆盖binding
			"ID":    "",
			"Title": "len>10",
			"Works": "len>2",
		},
		"delete": []string{
			"ID",
		},
	}
}

// 有自定义校验标签时，一定要写同名校验方法
func (t Example) CheckAge(rv reflect.Value) bool {
	if rv.String() == "1" {
		return true
	}
	return false
}

// 第二种写法
func (t Example) CheckAge2(vf validator.FieldLevel) bool {
	if vf.Field().String() == "1" {
		return true
	}
	return false
}

// 整体性校验，一些上面的方法校验不出来的时候，都可以在这个方法里校验
func (t Example) WholeCheck(scene string) map[string]string {
	warns := make(map[string]string)
	if scene == "delete" && t.ID != 1 {
		warns["id"] = "id不能等于1"
	}
	return warns
}
