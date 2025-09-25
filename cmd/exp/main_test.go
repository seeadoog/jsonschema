package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestJSON(t *testing.T) {

	je := json.NewDecoder(bytes.NewReader([]byte(`{} {} {}`)))
	for {
		var i any
		err := je.Decode(&i)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(i)
	}

}
