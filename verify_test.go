package verify

import (
	"fmt"
	"reflect"
	"testing"
)


type Dog struct {
	TT string `json:"tt" binding:"required"`
	TT2 string `json:"tt2" binding:"required"`
	Num int `json:"num" binding:"min=3,max=2"`
	DOO DD `json:"doo"`
	Formulas  []Formu `json:"formulas" binding:"gt=0,dive,required" byname:"公式"`
}

type Formu struct {
	Name string `json:"name"`
	Expr string `json:"expr" binding:"required"`
	Code string `json:"code" binding:"required" byname:"代码"`
}

type DD struct {
	Title      string `json:"title"`
	ReportDate string `json:"report_date" binding:"CheckID" byname:"日期"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
}

func (t Dog) Warns() map[string]string {
	return map[string]string{
		"required": "{0}是必须的",
		"CheckID": "{0}格式不对",
	}
}

func (t Dog) Scenes() map[string]interface{} {
	return map[string]interface{}{
		"create": []string{
			"TT",
		},
		"update": map[string]string{
			"TT":        "required",
		},
		"delete": []string{
			"TT2",
		},
	}
}

func (t Dog) CheckID(rv reflect.Value) bool {
	if rv.String() == "1" {
		return true
	}
	return false
}

func (t Dog) WholeCheck(scene string) map[string]string {
	warns := make(map[string]string)
	if t.TT == t.TT2 && scene == "create" {
		warns["tt"] = "tt不能等于tt2"
	}
	return warns
}


func TestScenes(t *testing.T) {
	var tt = Dog{
	}
	scenes := tt.Scenes()
	for _, val := range scenes {
		switch vv := val.(type) {
		case []string:
			fmt.Println(vv)
		case map[string][]string:
			fmt.Println(991)
		}
	}
}

func TestCheck(t *testing.T) {
	var tt = &Dog{
		//TT: "sstt",
		Formulas: []Formu{
			{Expr:"1"},
		},
	}
	pass, warns, err := Check(tt)
	fmt.Println(pass, warns, err)
}
