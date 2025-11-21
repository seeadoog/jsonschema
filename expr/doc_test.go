package expr

import (
	"testing"
)

func TestDoc2(t *testing.T) {
	showDocOf("ctx.", &Usr{})
	//fmt.Println(showDocOf("ctx.", &Usr{}))
}

func TestDOc3(t *testing.T) {

	//fmt.Println(showDocOf("", addV))
}

type V struct {
	typ     Type
	integer int
	str     string

	Name string
}

func (v V) Int() int {
	return v.integer
}

func (v V) String() string {
	return v.str
}

func addV(a V, b V) (r V) {
	r.integer = a.integer + b.integer
	return r
}

var (
	c V
)

func BenchmarkAddV2(bb *testing.B) {
	a := V{}
	b := V{}

	for i := 0; i < bb.N; i++ {

		c = addV(a, b)
	}
}
