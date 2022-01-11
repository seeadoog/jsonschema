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

func TestStruct(t *testing.T) {
	sc := `
{
	"type":"object"
}
`
	s := &Schema{}
	err := json.Unmarshal([]byte(sc), s)
	fmt.Println(err)
	type A struct {
	}
	i := A{}
	tt(i)

	fmt.Println(s.Validate(i))
}
func tt(i interface{}) {
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
			"format":"phone"
		}
	}
}
`
	rootSchema := Schema{}

	err := json.Unmarshal([]byte(schema), &rootSchema)
	if err != nil {
		panic(err)
	}

	js := `
{
	"name":"1"
}
`
	var o interface{}
	err = json.Unmarshal([]byte(js), &o)
	if err != nil {
		panic(err)
	}
	fmt.Println(rootSchema.Validate(o))
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

func TestArray(t *testing.T) {
	schema := `{
"type":"object",
"properties":{
	"app_id":{
		"type":"string"
	},
	"vcn":{
		
	},
    "ent":{},
	"ids":{
		"type":"array|string"
	}
},
"allOf":[
	{
		"if":{
			"keyMatch":{
				"app_id":"sms"
			}
		},
		"then":{
			"setVal":{
				
				"ids":{
					"func":"append",
					"args":["ent",",",{
					"func":"join",
					"args":[["1","2"],","]
				}]
				}
			}
		}
	}

]

}

`
	validate(schema,`{"app_id":"sms","ent":"x2","vcn":"xiaoyan","ids":[1,2,3,4,5]}`)
}
