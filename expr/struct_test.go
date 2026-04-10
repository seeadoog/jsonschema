package expr

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
	"testing"
	"unsafe"
)

type Usr struct {
	Name    string
	Age     int
	Friends []*Usr
	Bytes   []byte
	Object  *Usr
}
type User2 struct {
	*Usr
}

func (u *Usr) Add(b Usr) string {
	return u.Name + b.Name
}

func (u *Usr) AddP(b *Usr) string {
	return u.Name + b.Name
}

func (u *Usr) AddAge(b Usr) int {
	return u.Age + b.Age
}

func (u *Usr) AddFriends(v []*Usr) string {
	na := ""
	for _, v := range v {
		na += v.Name
	}
	return na
}

func (u *Usr) Joins(ss ...string) string {
	return strings.Join(ss, "")
}

func (u *Usr) Joins2(a string, ss ...string) string {
	return a + strings.Join(ss, "")
}

func (u *Usr) Return2(arr []string) (string, string) {
	return arr[0], arr[1]
}

func (u *Usr) ReturnE(arr []string) (string, error) {
	return arr[0], errors.New("ERR")
}

func (u *Usr) ReturnE2(arr []string) (string, error) {
	return arr[0], nil
}
func (u *Usr) PrintMap(m map[string]string) string {
	kv := []string{}
	for k, v := range m {
		kv = append(kv, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(kv)
	return strings.Join(kv, ",")
}

func (u *Usr) Ctx(c *Context, a string) string {
	return c.GetByString("test").(string) + a
}

func (u *Usr) Ctx2(c *Context, a ...string) string {
	return c.GetByString("test").(string) + strings.Join(a, "")
}

func (u *Usr) Opt(c *Context, a string, o *Options) string {
	return o.GetStringDef(a, "66")
}

func (u *Usr) Cb(c *Context, f func(a string) string, a string) string {
	return f(a)
}

func (u *Usr) Cb2(c *Context, f func(a string) (string, string), a string) string {
	a, b := f(a)
	return a + b
}

func TestStruct2(t *testing.T) {
	e, err := ParseFromJSONStr(`
[
"usr.Friends[0].Name='he'",
"a = usr->Name",
"c = u3.Add({Name: 'xx'})",
"d = u3.AddP({Name: 'xx'})",
"e = u3.AddAge({Age:100})",
"f = u3.AddFriends([{Name:'xx2'},{Name:'xx3'}])",
"g = u3.PrintMap({name:'a',age:6})",
"u3.Friends = [{Name:'a1',Age:90}]",
"u3.Bytes = 'hello'",
"u3.Object = {Name:'obj',Age:55}",
"u5.Usr.Age=30",
"h = u5.Joins(['1','2']...)",
"i = u5.Joins2('a',['1','2']...)",
"j = u5.Joins2('a')",
"k = u5.Return2(['22','33'])",
"l = u5.ReturnE(['22'])",
"m = u5.ReturnE2(['22'])",
"n = u5.Ctx('a')",
"o = u5.Ctx2(['a'])",
"p = u5.Opt('name',{name:'55'})",
"q = u5.Opt('name')",
"r = fmt.sprintf('a=%v',1)",
"s = u5.Cb(s => s+s,'aa')",
"t = u5.Cb2($ =>[$,$],'a2')"
]
`)
	if err != nil {
		t.Fatal(err)
	}

	u3 := &Usr{Name: "u3"}
	c := NewContext(map[string]any{
		"usr": &Usr{
			Name:    "Alice",
			Age:     0,
			Friends: nil,
		},
		"u3": u3,
		"u4": &Usr{Name: "u4"},
		"map": map[string]string{
			"a": "A",
			"b": "B",
		},
		"fmt": map[string]any{
			"sprintf": fmt.Sprintf,
		},
		"u5": &User2{
			Usr: &Usr{Name: "u5"},
		},
		"test": "test",
	})
	c.ForceType = false
	err = c.Exec(e)
	if err != nil {
		panic(err)
	}
	assertEqual(t, c, "a", "Alice")
	assertEqual(t, c, "c", "u3xx")
	assertEqual(t, c, "d", "u3xx")
	assertEqual(t, c, "e", 100)
	assertEqual(t, c, "f", "xx2xx3")
	assertEqual(t, c, "g", "age=6,name=a")
	assertEqual(t, c, "u3.Friends[0]", u3.Friends[0])
	assertEqual(t, c, "u3.Bytes.string()", ("hello"))
	assertEqual(t, c, "u3.Object.Name", ("obj"))
	assertEqual(t, c, "u3.Object.Age", 55)
	assertEqual(t, c, "u5.Name", "u5")
	assertEqual(t, c, "u5.Usr.Age", 30)
	assertEqual(t, c, "u5.Age", 30)
	assertEqual(t, c, "h", "12")
	assertEqual(t, c, "i", "a12")
	assertEqual(t, c, "j", "a")
	assertEqual(t, c, "k[0]", "22")
	assertEqual(t, c, "k[1]", "33")
	assertEqual(t, c, "l[0]", "22")
	assertEqual(t, c, "l[1].Error()", "ERR")
	assertEqual(t, c, "m[1]==nil", true)
	assertEqual(t, c, "n", "testa")
	assertEqual(t, c, "o", "testa")
	assertEqual(t, c, "p", "55")
	assertEqual(t, c, "q", "66")
	assertEqual(t, c, "r", "a=1")
	assertEqual(t, c, "s", "aaaa")
	assertEqual(t, c, "t", "a2a2")

}

func TestStruct(t *testing.T) {
	e, err := ParseFromJSONStr(`
[
"a = usr->Name",
"b = usr->Age",
"c = len(usr->Friends)",
"d = usr->Friends[0]->Name",
"usr2->Name = 'may'",
"usr2->Age = 3",
"usr2->Friends[0]->Age = 6",
"usr2->Friends[0]->Name = 'jk'",
"sum=0;for(arrs,e=>sum=sum+e)"
]
`)
	if err != nil {
		panic(err)
	}

	u := &Usr{
		Name: "bob",
		Age:  15,
		Friends: []*Usr{
			{
				Name: "tom", Age: 2,
			},
		},
	}
	u2 := &Usr{
		Name:    "",
		Friends: []*Usr{{}},
	}
	c := NewContext(map[string]any{
		"usr":  u,
		"usr2": u2,
		"arrs": []float64{1, 2, 3},
	})
	c.ForceType = false
	err = c.Exec(e)
	if err != nil {
		panic(err)
	}

	assertEqual(t, c, "a", u.Name)
	assertEqual(t, c, "b", (u.Age))
	assertEqual(t, c, "c", float64(len(u.Friends)))
	assertEqual(t, c, "d", u.Friends[0].Name)
	assertEqual(t, c, "sum", float64(6))
	assertEqual2(t, u2.Name, "may")
	assertEqual2(t, u2.Age, 3)
	assertEqual2(t, u2.Friends[0].Age, 6)
	assertEqual2(t, u2.Friends[0].Name, "jk")
}

func BenchmarkDtring(b *testing.B) {
	aa := "xxxx"
	var c string
	for i := 0; i < b.N; i++ {
		c = reflect.ValueOf(aa).String()
	}
	fmt.Println(c)
}

var Sink string
var Sink2 reflect.Value

func BenchmarkReflectString(b *testing.B) {
	aa := "hello world"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := reflect.ValueOf(aa)
		Sink2 = v         // 防止优化
		Sink = v.String() // 防止优化
	}
}

func TestHash5(t *testing.T) {
	m := make(map[uint64]string)
	sum := 0
	rangeString(m, make([]byte, 6), 0, &sum, 3)
}

func rangeString(m map[uint64]string, bf []byte, idx int, sum *int, n int) {
	if idx >= n {
		h := calcHash(ToString(bf[:n]))
		*sum++
		if m[h] != "" {
			if m[h] != ToString(bf[:n]) {
				panic("hash conflict" + ToString(bf[:n]) + ":" + m[h])
			}
		}
		m[h] = ToString(bf[:n])

		return
	}
	for i := 'A'; i <= 'z'; i++ {
		bf[idx] = byte(i)
		rangeString(m, bf, idx+1, sum, n)
	}
	for i := '0'; i <= '9'; i++ {
		bf[idx] = byte(i)
		rangeString(m, bf, idx+1, sum, n)
	}
}

func BenchmarkHashMM(b *testing.B) {
	for i := 0; i < b.N; i++ {
		calcHash("xxx")
	}
}

func TestValue(t *testing.T) {
	a := int64(-1)
	b := uint64(a)
	c := int64(b)

	fmt.Println(a, b, c)

}

var (
	_v any

	_b = "data"
)

func BenchmarkValue(b *testing.B) {
	v := Value{}
	b.ReportAllocs()
	v.SetFloat(1)

	bb := v
	for i := 0; i < b.N; i++ {

		v.Equal(&bb)
	}
}

type VV struct {
	king int
	str  string
	num  int
}

func (v VV) Equal(b VV) bool {
	return v == b
}

type Kind int8

const (
	KindInt Kind = iota
	KindFloat
	KindString
	KindByte
	KindAny
)

type Value struct {
	sl   int64
	p1   unsafe.Pointer
	p2   int64
	any  any
	kind Kind
}
type stringHeader struct {
	data unsafe.Pointer
	len  int64
}

type sliceHeader struct {
	data unsafe.Pointer
	len  int64
	cap  int64
}

type interf struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}

func (v *Value) SetAny(n any) {
	v.any = n
	v.kind = KindAny
}

func (v Value) Any() any {
	if v.kind != KindAny {
		panic(fmt.Sprintf("fail to use kind '%v' as kind any", v.kind))
	}
	return v.any
}

func (a *Value) Equal(b *Value) bool {
	if a.kind != b.kind {
		return false
	}
	switch a.kind {
	case KindString:
		return a.String() == b.String()
	case KindInt:
		return a.Int() == b.Int()
	case KindFloat:
		return a.Float() == b.Float()
	case KindByte:
		return bytes.Equal(a.Bytes(), b.Bytes())
	case KindAny:
		return a.any == b.any
	}
	return false
}

func (v *Value) String() string {
	if v.kind != KindString {
		panic(fmt.Sprintf("fail to use kind '%v' as kind string", v.kind))
	}
	sh := stringHeader{
		data: v.p1,
		len:  v.sl,
	}
	return *(*string)(unsafe.Pointer(&sh))
}
func (v *Value) SetString(s string) {
	sh := (*stringHeader)(unsafe.Pointer(&s))
	v.p1 = sh.data
	v.sl = sh.len
	v.kind = KindString
}

func (v *Value) Int() int64 {
	if v.kind != KindInt {
		panic(fmt.Sprintf("fail to use kind '%v' as kind int", v.kind))
	}
	return (v.sl)
}

func (v *Value) Bytes() []byte {
	if v.kind != KindByte {
		panic(fmt.Sprintf("fail to use kind '%v' as kind []byte", v.kind))
	}
	sh := sliceHeader{
		data: v.p1,
		len:  v.sl,
		cap:  v.p2,
	}
	return *(*[]byte)(unsafe.Pointer(&sh))
}

func (v *Value) SetBytes(s []byte) {
	sh := (*sliceHeader)(unsafe.Pointer(&s))
	v.p1 = sh.data
	v.sl = sh.len
	v.p2 = sh.cap

	v.kind = KindByte
}

func (v *Value) SetInt(n int64) {
	v.sl = n
	v.kind = KindInt
}

func (v *Value) SetFloat(n float64) {
	v.sl = *(*int64)(unsafe.Pointer(&n))
	v.kind = KindFloat
}

func (v *Value) Float() float64 {
	if v.kind != KindFloat {
		panic(fmt.Sprintf("fail to use kind '%v' as kind float", v.kind))
	}
	return *(*float64)(unsafe.Pointer(&v.sl))
}

func TestUnsafeValueStringGC(t *testing.T) {
	for i := 0; i < 100000; i++ {
		s := strings.Repeat("A", 1024) // 分配在堆上
		v := &Value{}
		v.SetString(s)

		// 切断所有 Go 级别的强引用
		s = ""

		// 制造 GC 压力
		for j := 0; j < 10; j++ {
			_ = make([]byte, 1024*2)
		}
		runtime.GC()

		out := v.String()

		if len(out) != 1024 {
			t.Fatalf("length corrupted: %d", len(out))
		}

		for _, c := range out {
			if c != 'A' {
				t.Fatalf("data corrupted: %q", out[:16])
			}
		}
	}
}

func TestUnsafeValueBytesGC(t *testing.T) {
	for i := 0; i < 10000; i++ {
		b := make([]byte, 1024)
		for i := range b {
			b[i] = 'A'
		}

		v := &Value{}
		v.SetBytes(b)

		b = nil

		for j := 0; j < 1000; j++ {
			_ = make([]byte, 1024)
		}

		runtime.GC()

		out := v.Bytes()
		for _, c := range out {
			if c != 'A' {
				t.Fatalf("corrupted")
			}
		}
	}
}

func TestSizeof(t *testing.T) {
	fmt.Println(unsafe.Sizeof(Value{}))
}

type vvaal interface {
	Val(c *Context) Value
}

type fvalfunc func(c *Context, vs ...vvaal) Value

func addfv(c *Context, vs ...vvaal) (nv Value) {
	a := vs[0].Val(c)
	b := vs[1].Val(c)
	nv.SetInt(a.Int() + b.Int())
	return nv
}

type constValue struct {
	v Value
}

func (c2 *constValue) Val(c *Context) Value {
	return c2.v
}

func BenchmarkValcx(b *testing.B) {
	a := Value{}
	a.SetInt(100)
	c := a
	ctx := &Context{}

	b.ReportAllocs()
	av := &constValue{a}

	bv := &constValue{c}
	f, _ := os.Create("test.pprof")
	defer f.Close()
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	b.ResetTimer()
	//args := []vvaal{av, bv}
	for i := 0; i < b.N; i++ {
		//v1 := av.Val(ctx)
		//v2 := bv.Val(ctx)
		//cv := Value{}
		//cv.SetInt(v1.Int() + v2.Int())
		addFv(ctx, av, bv)
		//addfv(ctx, args...)
	}
}

func addFv(ctx *Context, av, bv vvaal) (cv Value) {
	v1 := av.Val(ctx)
	v2 := bv.Val(ctx)
	cv.SetInt(v1.Int() + v2.Int())
	return cv
}

var (
	arrs = make(map[any]any, 10)
)

func BenchmarkArrs(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		arrs["xx"] = 4
	}
}

func BenchmarkMapPath(b *testing.B) {
	m := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": 5,
			},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m["a"].(map[string]any)["b"].(map[string]any)["c"]
	}
}

type valstruct struct {
	i float64
}

func (v valstruct) Val(c *Context) valstruct {
	return v
}

type valinter interface {
	Val(c *Context) valstruct
}

type addValStruct struct {
	a valinter
	b valinter
}

func (a *addValStruct) Val(c *Context) valstruct {
	return valstruct{
		i: a.a.Val(c).i + a.b.Val(c).i,
	}

}

func BenchmarkValInter(b *testing.B) {
	av := &addValStruct{
		a: valstruct{},
		b: valstruct{},
	}
	for i := 0; i < b.N; i++ {

		av.Val(nil)

	}
}
