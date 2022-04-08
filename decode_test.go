package jsonschema

import (
	"encoding/json"
	"fmt"
	"testing"
)

type User struct {
	Name   string                 `json:"name"`
	Age    int                    `json:"age"`
	Sister map[string]interface{} `json:"sister"`
	Childs [2]*User               `json:"childs"`
}

func (s *User) String() string {
	bs, _ := json.Marshal(s)
	return string(bs)
}

func TestUnmarshalMap(t *testing.T) {

	m := map[string]interface{}{
		"name": "lixiang",
		"age":  5,
		"sister": map[string]interface{}{
			"name": "mary",
			"age":  6,
		},
		"childs": []interface{}{
			map[string]interface{}{
				"name": "jhon",
				"age":  3,
			},
		},
	}

	var v User
	err := UnmarshalFromMap(m, &v)
	if err != nil {
		panic(err)
	}
	fmt.Println(v)
	//fmt.Println(v == interface{}(m))
}

func BenchmarkName(b *testing.B) {
	j := []byte(`{"name":"lixiang","age":5,"sister":{"age":6,"name":"mary"},"childs":[{"name":"jhon","age":3,"sister":null,"childs":null}]}`)
	var r User
	json.Unmarshal(j, &r)
	fmt.Println(r)
	for i := 0; i < b.N; i++ {
		var r User
		json.Unmarshal(j, &r)

	}
}

func BenchmarkName2(b *testing.B) {
	j := []byte(`{"name":"lixiang","age":5,"sister":{"age":6,"name":"mary"},"childs":[{"name":"jhon","age":3,"sister":null,"childs":null}]}`)

	for i := 0; i < b.N; i++ {
		var m interface{}
		json.Unmarshal(j, &m)
		var r User
		UnmarshalFromMap(m, &r)

	}
}

func TestJ2(t *testing.T) {
	j := []byte(`{"name":"lixiang","age":5,"sister":{"age":6,"name":"mary"},"childs":[{"name":"jhon","age":3,"sister":null,"childs":null}]}`)

	var m interface{}
	json.Unmarshal(j, &m)
	var r User
	UnmarshalFromMap(m, &r)

	fmt.Println(r)
}
