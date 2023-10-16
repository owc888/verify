# 参数验证
封装 [github.com/go-playground/validator/v10](github.com/go-playground/validator/v10) 验证器，以简化参数的验证。
## 使用方法
1.引入

    import "github.com/owc888/verify"

2.编写需要校验数据的结构体。在tag中，binding填写校验规则（具体校验规则可查看[validator package](https://pkg.go.dev/github.com/go-playground/validator/v10)）；byname写参数别名，即输出时，用byname替换参数名，原校验器对参数名的显示并不友好。

    type Example struct {
        ID     int    `json:"id" binding:"required"`
        Title  string `json:"title" binding:"required" byname:"标题"`
        Author Author `json:"author"`
        Works  []Work `json:"works" binding:"gt=0,dive,required" byname:"公式"`
    }

    type Author struct {
        FirstName string `json:"first_name" byname:"姓"`
        LastName  string `json:"last_name"  byname:"名"`
        Age       string `json:"age" binding:"CheckAge" byname:"年龄"`
    }

    type Work struct {
        Name string `json:"name" byname:"名字"`
        Code string `json:"code" binding:"required" byname:"代码"`
    }

3.个性化校验，在binding中写入自定义的校验方法名，同时编写好校验方法，有两种写法，主要是传参类型不同。沿用`github.com/go-playground/validator/v10`的功能。

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

4.可对结构体附加参数可设置在不同场景下，针对不同的参数进行校验。添加如下方法并返回正确的数据格式的数据即可。

    func (t Example) Scenes() map[string]interface{} {
        return map[string]interface{}{
            "create": []string{
                "Title", "Works",
            },
            "update": map[string]string{
                "ID","Title":,"Works",
            },
            "delete": []string{
                "ID",
            },
        }
    }

4.整体性的校验，当以上校验都不能满足要求时，可以在以下方法进行校验

    // 整体性校验，一些上面的方法校验不出来的时候，都可以在这个方法里校验
    func (t Example) WholeCheck(scene string) map[string]string {
        warns := make(map[string]string)
        if scene == "delete" && t.ID != 1 {
            warns["id"] = "id不能等于1"
        }
        if t.Title == t.Author {
            warns["title"] = "标题不能等于作者名"
        }            
        return warns
    }
自定义方法可以自己构造方法名，场景设置和整体校验必须用Scenes()和WholeCheck()方法，不需要时可以直接不写方法。

使用时，将需要校验的参数传入Check()方法即可：

    var e = &Example{
		Title: "怎么验证",
	}
	pass, warns, err := Check(e, "update", "zh")

具体使用看查看测试用例。

## 未来更新计划
1.校验校验方法可传递额外参数

2.对于相同的校验规则，Warns()尝试对不同层级的不同字段有不同的告警信息，如'x1.x2:required' 提示”我需要x2“；同时'x1.x3:required'可提示”不能没有x3啊“，两者有不同的告警信息

3.针对不同场景的校验，可替换结构体里的binging，以覆盖原有的校验规则，进一步实现个性化

4.全局初始化验证工具，将tag信息保存下来，减少反射逻辑的执行



