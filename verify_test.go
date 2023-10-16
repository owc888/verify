package verify

import (
	"fmt"
	"testing"
)

func TestCheck(t *testing.T) {
	var e = &Example{
		Title: "怎么验证",
	}
	pass, warns, err := Check(e, "update", "zh")
	fmt.Println(pass, warns, err)
}
