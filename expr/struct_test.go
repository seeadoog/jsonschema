package expr

import (
	"testing"
)

type Usr struct {
	Name    string
	Age     int
	Friends []*Usr
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
"usr2->Friends[0]->Name = 'jk'"
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
	assertEqual2(t, u2.Name, "may")
	assertEqual2(t, u2.Age, 3)
	assertEqual2(t, u2.Friends[0].Age, 6)
	assertEqual2(t, u2.Friends[0].Name, "jk")
}
