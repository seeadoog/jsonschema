package expr

import "testing"

//
//import (
//	"fmt"
//	"testing"
//)
//
//var (
//	testExpr = `
//$.name = 1
//`
//	ctx = NewContext(map[string]interface{}{
//		"a": "hello",
//		"d": "arch",
//		"mm": map[string]any{
//			//"aaa": "1",
//			//"bbb": "2",
//			"sub": []any{1, 2, 3, 4, 5, 6},
//		},
//	})
//)
//
//func init() {
//	RegisterFunc("test2", func(ctx *Context, args ...Val) any {
//		return nil
//	}, 2)
//	//RegisterFunc("print", func(ctx *Context, args ...Val) any {
//	//	return nil
//	//}, -1)
//
//}
//func TestLambda(t *testing.T) {
//	e, err := ParseValue(testExpr)
//	if err != nil {
//		panic(err)
//	}
//
//	ec, err := compileRootValue(e)
//	if err != nil {
//		panic(err)
//	}
//
//	ctx.SetStackAndEnv(ec, "$", map[string]any{
//		"bb": 1,
//	})
//
//	ctx.InitStackSize(ec.StackSize() * 2)
//
//	fmt.Println(ec.Val(ctx))
//	//fmt.Println(ec.Val(ctx))
//	//fmt.Println(ec.Val(ctx))
//	fmt.Println(ctx.table)
//	fmt.Println(ctx.stack)
//	fmt.Println("stack:", ctx.GetFromStack(ec, "b"))
//
//}
//
//func BenchmarkCompiledValue(b *testing.B) {
//	e, err := ParseValue(testExpr)
//	if err != nil {
//		panic(err)
//	}
//
//	ec, err := compileRootValue(e)
//	if err != nil {
//		panic(err)
//	}
//
//	fmt.Println(ec.Val(ctx))
//	b.ReportAllocs()
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		e.Val(ctx)
//	}
//}
//
//func BenchmarkCompiledValueRaw(b *testing.B) {
//	e, err := ParseValue(testExpr)
//	if err != nil {
//		panic(err)
//	}
//	ctx.NewCallEnv = false
//
//	fmt.Println(e.Val(ctx))
//	b.ReportAllocs()
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		e.Val(ctx)
//	}
//}

func BenchmarkCol(b *testing.B) {
	c := NewContext(map[string]any{})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.Clone()
	}
}
