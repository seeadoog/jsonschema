package expr

import (
	"fmt"
	"testing"
)

func TestDoc2(t *testing.T) {

	s := showDocOf("ctx.", &Usr{})
	fmt.Println(s)
}
