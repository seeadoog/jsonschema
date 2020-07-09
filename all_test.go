package jsonschema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func validate(schema, js string) {
	sc := &Schema{}
	if err := json.Unmarshal([]byte(schema), sc); err != nil {
		panic(err)
	}
	var i interface{}
	if err := json.Unmarshal([]byte(js), &i); err != nil {
		panic(err)
	}

	if err := sc.Validate(i); err != nil {
		fmt.Println(err)
	}
	b, _ := json.Marshal(i)
	fmt.Println("after=>", string(b))
}

func TestStruct(t *testing.T){
	sc:=`
{
	"type":"object"
}
`
	s:=&Schema{}
	err:=json.Unmarshal([]byte(sc),s)
	fmt.Println(err)
	type A struct {

	}
	i:=A{}
	tt(i)

	fmt.Println(s.Validate(i))
}
func tt(i interface{}){
	switch i.(type) {
	case struct{}:
		fmt.Println("----")
	default:
		fmt.Println(reflect.TypeOf(i))
	}
}
func TestBase(t *testing.T) {
	schema := `
{
	"type":"object",
	"properties":{
		"name":{
			"type":"string|number",
			"maxLength":5,
			"minLength":1,
			"maximum":10,
			"minimum":1,
			"enum":["1","2"],
			"replaceKey":"name2",
			"formatVal":"string",
			"format":"phone1"
		}
	}
}
`

	js := `
{
	"name":"15029332345"
}
`
	validate(schema, js)
}

func TestMagic(t *testing.T) {
	schema := `
{
  "type": "object",
  "switch": "name",
  "case": {
    "jhon": {
        "setVal": {
          "all_name": {
            "func": "append",
            "args": ["${name}","_","${age}"]
          }
      }
    },
    "alen": {
      "required": ["age"]
    }
  },
  "if":{
		"keyMatch":{
			"name":"jhon",
			"age":5
		}
	},
	"then":{
		"required":["age2"],
		"setVal":{
			"name_coy":"${name}"
		}
	}
}
`
	validate(schema, `
{
	"name":"jhon",
	"age":5
}

`)
}
