package jsonpath

import (
	"encoding/json"
	"fmt"
	"testing"
	"unsafe"
)

func TestJPATH(t *testing.T) {
	c := Complied{
		indexes: []index{
			indexMap("arr"), newIndexSlice(1), newIndexSlice(0),
		},
	}
	m := map[string]interface{}{
		"name": "jim",
		"chd": map[string]interface{}{
			"name": "xxx",
		},
		"arr": []any{
			"arr1",
			[]any{"arr11"},
		},
	}
	fmt.Println(c.Get(m))

}

func TestJPATHSet(t *testing.T) {
	c := Complied{
		indexes: []index{
			indexMap("chd"), newIndexSlice(1), newIndexSlice(3), indexMap("name"), newIndexSlice(0),
		},
	}
	m := map[string]interface{}{
		//"name": "jim",
		//"chd": map[string]interface{}{
		//	"name": "xxx",
		//},
	}
	fmt.Println(c.Set(m, 1))

	ns, _ := json.MarshalIndent(m, "", "    ")
	fmt.Println(string(ns))

}

func TestCompile(t *testing.T) {
	jp, err := compileExpr("abb\\.\\.55.[4][5][6].aa\\[5")
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(jp.indexes)
	fmt.Println(string(data))
}

func TestJset(t *testing.T) {
	var m any
	setVal2("[0].user[0].a.b", &m, 1)
	must(setVal2("[1].user[1].a.b", &m, 1))
	//setVal("user[-1].a", m, 1)
	//setVal("user[-1].a", m, 1)

	ns, _ := json.MarshalIndent(m, "", "    ")
	fmt.Println(string(ns))

}

func TestSet2(t *testing.T) {
	m := map[string]interface{}{}

	must(setVal("ws[-1].cw[-1]", m, 1))
	must(setVal("ws[-1].cw[-1]", m, 1))
	must(setVal("ws[-1].cw[-1]", m, 1))
	must(setVal("ws[-1].cw[-1]", m, 1))
	//must(setVal("ws[-1][-1][-1]", m, 3))
	//must(setVal("ws[1][1][1].a", m, 1))
	//setVal("a.b.d", m, 1)
	//setVal("a.b.e", m, 1)
	//setVal("a.b.f[3]", m, 1)
	//setVal("a.b.f[4]", m, 1)

	ns, _ := json.MarshalIndent(m, "", "    ")
	fmt.Println(string(ns))
}

func setVal(expr string, src, val any) error {
	jp, err := Compile(expr)
	if err != nil {
		return err
	}
	return jp.Set(src, val)
}

func setVal2(expr string, src, val any) error {
	jp, err := Compile(expr)
	if err != nil {
		return err
	}
	return jp.Set2(src, val)
}

func updateSliceInter(in any, new []any) {
	type face struct {
		t, d unsafe.Pointer
	}
	ptr := (*face)(unsafe.Pointer(&in))

	pp := (*[]any)(ptr.d)
	*pp = new
}

func TestUpdateS(t *testing.T) {
	ss := []any{1}
	updateSliceInter(ss, []any{1, 2, 3})
	fmt.Println(ss)
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}

func TestSet22(t *testing.T) {

	var src any
	must(setVal2("c.name[0]", &src, 1))
	must(setVal2("c.age[0]", &src, 1))
	must(setVal2("c.age[1]", &src, 1))

	bs, _ := json.MarshalIndent(src, "", "\t")
	fmt.Println(string(bs))
}

func TestJSON(t *testing.T) {
	type tmp struct {
		Value *Complied `json:"value"`
	}
	v := &tmp{}
	err := json.Unmarshal([]byte(`{"value":"namage.xx"}`), v)
	if err != nil {
		panic(err)
	}
	fmt.Println(v.Value)
	// a = 5
	// b = 5
	// print(a,b)
}
