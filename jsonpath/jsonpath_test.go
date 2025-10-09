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
	must(setVal2("c.name", &src, "ase"))
	must(setVal2("c.age", &src, 1))
	must(setVal2("c.age", &src, 1))
	must(setVal2("c.${name}.dd", &src, "from_ase"))

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

func BenchmarkJPs(b *testing.B) {
	jp, err := Compile("a[0][1][2][3]")
	if err != nil {
		panic(err)
	}
	b.ReportAllocs()
	v := map[string]interface{}{}
	for i := 0; i < b.N; i++ {
		jp.Set(v, 1)
	}
	fmt.Println(v)

	bs, _ := json.MarshalIndent(v, "", "\t")
	fmt.Println(string(bs))
}

var (
	username = MustCompile("o1.o2.o3.o4")
	age      = MustCompile("common.age")
)

func TestF(t *testing.T) {
	var o any

	err := json.Unmarshal([]byte(`{"common": {
	"name":"usname",
	"age":4,
	"o2":{
		"o3":{
			"o4":{}
		}
	}

}}`), &o)
	if err != nil {
		panic(err)
	}

	fmt.Println(username.GetStringDef(o, "[0]"))
	fmt.Println(age.GetNumberDef(o, 12))
}

func BenchmarkName(b *testing.B) {
	var o any

	err := json.Unmarshal([]byte(`
{
 "datad": "1",
 "doc": {},
 "o1": {
  "arr": [
   1,
   2.2,
   3.3
  ],
  "data": "hello",
  "o2": {
   "o3": {
    "o4": 1
   },
   "xx": 1
  },
  "text": "js is ok",
  "text2": "js is ok",
  "text3": "js is ok",
  "text4": "js is ok"
 },
 "res": {
  "err": "err",
  "data": "data"
 },
 "status": 3,
 "usr": {
  "Name": "55",
  "Age": 18,
  "Chd": {
   "Name": "chd",
   "Age": 3,
   "Chd": null,
   "Arr": null
  },
  "Arr": null
 }
}


`), &o)
	if err != nil {
		panic(err)
	}
	b.ReportAllocs()

	var v string
	for i := 0; i < b.N; i++ {
		username.Get(o)
		//v = o.(map[string]interface{})["common"].(map[string]any)["name"].(string)

	}
	fmt.Println(v)
}

func swh(i int) int {
	switch i {
	case 0:
		return 1
	case 1:
		return 2
	case 2:
		return swh(i - 1)
	}
	return swh(i - 1)
}
func TestSwi(t *testing.T) {

	swh(2)
}
