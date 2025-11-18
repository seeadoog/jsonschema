package expr

import (
	"fmt"
	"testing"
)

type Usr struct {
	Name    string
	Age     int
	Friends []*Usr
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

func TestStruct2(t *testing.T) {
	e, err := ParseFromJSONStr(`
[
"usr.Friends[0].Name='he'",
"a = usr->Name",
"c = u3.Add({Name: 'xx'})",
"d = u3.AddP({Name: 'xx'})",
"e = u3.AddAge({Age:100})",
"f = u3.AddFriends([{Name:'xx2'},{Name:'xx3'}])"
]
`)
	if err != nil {
		t.Fatal(err)
	}

	c := NewContext(map[string]any{
		"usr": &Usr{
			Name:    "Alice",
			Age:     0,
			Friends: nil,
		},
		"u3": &Usr{Name: "u3"},
		"u4": &Usr{Name: "u4"},
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
	assertEqual(t, c, "b", float64(u.Age))
	assertEqual(t, c, "c", float64(len(u.Friends)))
	assertEqual(t, c, "d", u.Friends[0].Name)
	assertEqual(t, c, "sum", float64(6))
	assertEqual2(t, u2.Name, "may")
	assertEqual2(t, u2.Age, 3)
	assertEqual2(t, u2.Friends[0].Age, 6)
	assertEqual2(t, u2.Friends[0].Name, "jk")
}

func TestHashType(t *testing.T) {
	fmt.Println(9 & 7)
	fmt.Println(17 & 7)

	//res = http_request().catch()
	//res.err?
}
