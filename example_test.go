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
		
		"username":"root",
		"class":"8",
		"age":37,
		"cip":"1.2.3.45",
		"params":{
			"sms":"haha",
			"sad":3
		}
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

func BenchmarkExa(b *testing.B) {
	SetFunc("redis.get", NewFunc1(func(a1 string) any {
		return a1
	}))
	b.ReportAllocs()
	bs, err := ioutil.ReadFile("example.json")
	if err != nil {
		panic(err)
	}
	schema, err := NewSchemaFromJSON(bs)
	if err != nil {
		panic(err)
	}
	js := `{
		
		"username":"root",
		"class":"8",
		"age":37,
		"cip":"1.2.3.45",
		"params":{
			"sms":"haha",
			"sad":3
		}
	}`

	var obj any

	err = json.Unmarshal([]byte(js), &obj)
	if err != nil {
		panic(err)
	}

	for i := 0; i < b.N; i++ {
		schema.Validate(obj)
	}
}
