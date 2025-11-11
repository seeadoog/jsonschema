package main

import (
	"fmt"
	expr2 "github.com/seeadoog/jsonschema/v2/expr"
	"time"
)

func main() {
	//expr2.RegisterDynamicFunc("get_cur_time", 0)
	expr, err := expr2.ParseValue(`'${name}_${get_cur_time()::std()}'`)
	if err != nil {
		panic(err)
	}
	ctx := expr2.NewContext(map[string]any{
		"name": "hello",
	})
	//ctx.SetFunc("get_cur_time", func(ctx *expr2.Context, args ...expr2.Val) any {
	//	return time.Now()
	//})

	expr2.SelfDefine0("std", func(ctx *expr2.Context, self time.Time) string {
		return self.Format("2006-01-02 15:04:05")
	})

	n := expr.Val(ctx)
	fmt.Println("result is:", n)

}
