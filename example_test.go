package jsonschema

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestExample(t *testing.T) {

	SetFunc("redis.get", NewFunc1(func(a1 string) any {
		return a1
	}))

	bs, err := ioutil.ReadFile("example.json")
	if err != nil {
		t.Fatal(err)
	}
	schema, err := NewSchemaFromJSON(bs)
	if err != nil {
		t.Fatal(err)
	}
	js := `{
		"name":"root",
		"age":24,
		"client_ip":"10.2.2.2",
		"names":["bob","johbn"],
		"js":"{}",
		"key":"key",
		"hd":{
			"name":"key"
		},
		"class":5
	}`

	var obj any

	err = json.Unmarshal([]byte(js), &obj)
	if err != nil {
		t.Fatal(err)
	}

	err = schema.Validate(obj)
	if err != nil {
		fmt.Println(err)
	}
	res, _ := json.MarshalIndent(obj, "", "\t")
	t.Log(string(res))

}
