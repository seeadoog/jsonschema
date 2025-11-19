package expr

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
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
"h = u5.Joins(['1','2'])",
"i = u5.Joins2('a',['1','2'])",
"j = u5.Joins2('a')",
"k = u5.Return2(['22','33'])",
"l = u5.ReturnE(['22'])",
"m = u5.ReturnE2(['22'])",
"n = u5.Ctx('a')",
"o = u5.Ctx2(['a'])"
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
