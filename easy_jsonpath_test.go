package jsonschema

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestJpsth(t *testing.T) {

	m := map[string]interface{}{
		//"tkn": []any{},
	}

	must(setJP("name[0]", m, 1))
	must(setJP("name[1]", m, 2))
	must(setJP("name[2].a", m, 3))
	must(setJP("name[2].b", m, 3))
	must(setJP("name[2].c.d", m, 3))
	must(setJP("name[2].c.f", m, 3))
	fmt.Println(m)

	bs, _ := json.MarshalIndent(m, "", "\t")
	fmt.Println(string(bs))

}

func setJP(path string, src, value interface{}) error {
	tk, err := parseJpathCompiled(path)
	if err != nil {
		panic(err)
	}
	return tk.Set(src, value)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func BenchmarkSetJP(b *testing.B) {

	b.ReportAllocs()
	jp, err := parseJpathCompiled("name")
	must(err)
	src := map[string]interface{}{}
	for i := 0; i < b.N; i++ {
		must(jp.Set(src, &i))
	}

}

func TestParseExpr(t *testing.T) {

}

type stack struct {
}

func (s *stack) Push(v interface{}) {}
func (s *stack) Pop() interface{} {
	return nil
}

type context struct {
	s *stack
}

type ff struct {
	InArgs int
	f      func(...any) any
}

func initArgs(f *ff) []any {

}

func (ff *ff) call(s *stack, args []any) {
	s.Pop()
}
