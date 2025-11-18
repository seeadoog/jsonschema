package expr

import (
	"fmt"
	"strings"
	"testing"
)

func execPost(code []binaryCode, ctx *Context, ss *stack[any]) any {

	for _, b := range code {
		if b.op == 0 {
			ss.push(b.val.Val(ctx))
		} else {
			switch b.op {
			case '&':
				l := ss.pop()
				r := ss.pop()
				if !l.(bool) {
					ss.push(false)
					continue
				}
				if !r.(bool) {
					ss.push(false)
					continue
				}
				ss.push(true)
			case '=':
				ss.push(ss.pop() == ss.pop())
			default:
				panic("unreachable op:" + string(byte(b.op)))
			}
		}
	}
	return ss.pop()
}

func TestPost(t *testing.T) {

	e, err := ParseValue(" a==1 && b == 3")
	if err != nil {
		t.Fatal(err)
	}
	br := toPost(e)
	//ss := stack[any]{}
	ctx := NewContext(nil)
	ctx.SetByString("a", 1.0)
	ctx.SetByString("b", 3.0)
	st := &stack[any]{}
	fmt.Println(execPost(br, ctx, st))

}

func BenchmarkPost2(b *testing.B) {
	v := strings.Repeat("1==1 &&", 0)
	e, err := ParseValue(v + "1==0 && 1==1 && 1==1 && 1 == 1 ")
	if err != nil {
		b.Fatal(err)
	}
	br := toPost(e)
	//ss := stack[any]{}
	ctx := NewContext(nil)
	ctx.SetByString("a", 1.0)
	ctx.SetByString("b", 3.0)
	st := &stack[any]{}
	fmt.Println(execPost(br, ctx, st))

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		execPost(br, ctx, st)
	}
}
func BenchmarkPost3(b *testing.B) {
	v := strings.Repeat("1==1 && ", 0)
	e, err := ParseValue(v + "1==1 || 1==1 || 1==1 || 1 == 1 ")
	if err != nil {
		b.Fatal(err)
	}
	//ss := stack[any]{}
	ctx := NewContext(nil)
	ctx.SetByString("a", 1.0)
	ctx.SetByString("b", 3.0)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		e.Val(ctx)
	}
}

func sw(i int) int {
	switch i {
	case 0:
		return 1
	case 2:
		return 3
	case 1:
		return 2
	case 3:
		return 5
	case 4:
		return 3
	case 5:
		return 0
	case 6:
		return 7
	case 7:
		return 8
	case 8:
		return 7

	case 9:
		return 1
	default:
		panic("xx")
	}
}

// kl,k
func BenchmarkSw(b *testing.B) {
	var k int
	for i := 0; i < b.N; i++ {
		k = sw(i & 7)
	}
	fmt.Println(k)
}
