package jsonschema

import (
	"encoding/json"
	"fmt"
	gjson "github.com/tidwall/gjson"
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

type Ids struct {
	Name int `json:"name"`
}

type Object struct {
	Ids []Ids `json:"ids2"`
}

func TestArray(t *testing.T) {
	schema := `{
"type":"object",
"properties":{
	"app_id":{
		"type":"string"
	},
	"d2":{
		"pattern":"^[0-9]{1,10}$"
	},
    "d":{},
	"ids":{
		"type":"array|string|integer"
	},
	"ids2":{
		"type":"array",
		"items":{
			"type":"object",
			"properties":{
				"name":{
					"type":"string"
				}
			}
		}
	},
	"time":{
		"type":"integer",
		"if":{
			"not":{
				"maximum":10000,
				"minimum":100
			}
		},
		"then":{
			"error":{
					"func":"sprintf",
					"args":["ddd is %v :%s","${$}","not easy"]
			}
		}
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
			"required":["d2"],
			"setVal":{
				
				"d":{
					"func":"or",
					"args":["${d}","defaultVal"]
                  },
				"ids":["$sprintf","appid is %v %v","${app_id}",["$join",[1,2,3,4,5],","]]
			}
		}
	}

]

}


	`
	sc := &Schema{}
	err := json.Unmarshal([]byte(schema), sc)
	if err != nil {
		panic(err)
	}
	obj := &Object{
		Ids: []Ids{{Name: int(5)}},
	}

	err = sc.Validate(obj)
	if err != nil {
		panic(err)
	}
}

func TestSchema(t *testing.T) {
	data := `
{
  "type":"object",
 
 
  "if":{
		"properties":{
			"param":{
				"not":{
					"pattern":"^[0-9]+$"
				}
				
			}
		}
	},
	"then":{
		"delete":["param"]
	}
}
`
	validate(data, `{"param":"50v","g":"50"}`)

}

func TestName(t *testing.T) {
	p := gjson.Parse(`
[{
  "type":"object",
 
 
  "if":{
		"properties":{
			"param":{
				"not":{
					"pattern":"^[0-9]+$"
				}
				
			}
		}
	},
	"then":{
		"delete":["param"]
	}
}]
`)
	p.ForEach(func(key, value gjson.Result) bool {
		fmt.Println(key, value.Type, value.IsArray())
		return true
	})
	fmt.Println(p.Type, p.IsArray())

}

func parseGjsonValue(r *gjson.Result) interface{} {
	switch r.Type {
	case gjson.String:
		return r.Str
	case gjson.Number:
		return r.Num
	case gjson.True:
		return true
	case gjson.False:
		return false
	case gjson.Null:
		return nil
	case gjson.JSON:
		if r.IsArray() {
			res := make([]interface{}, 0)
			r.ForEach(func(key, value gjson.Result) bool {
				res = append(res, parseGjsonValue(&value))
				return true
			})
			return res
		}
		if r.IsObject() {
			res := make(map[string]interface{})
			r.ForEach(func(key, value gjson.Result) bool {
				res[key.Str] = parseGjsonValue(&value)
				return true
			})
			return res
		}
	}
	return nil
}

func TestParseJ(t *testing.T) {

	p := gjson.Parse(`
[{
  "type":"object",
 
 
  "if":{
		"properties":{
			"param":{
				"not":{
					"pattern":"^[0-9]+$"
				}
				
			}
		}
	},
	"then":{
		"delete":["param"]
	}
}]
`)

	fmt.Println(parseGjsonValue(&p))
}

var (
	jsonstr = `
[{
  "type":"object",
 
 
  "if":{
		"properties":{
			"param":{
				"not":{
					"pattern":"^[0-9]+$"
				}
				
			}
		}
	},
	"then":{
		"delete":["param"]
	}
}]
`
)

func BenchmarkGJSON(b *testing.B) {

	for i := 0; i < b.N; i++ {
		p := gjson.Parse(jsonstr)

		parseGjsonValue(&p)
	}
}

func BenchmarkSTD(b *testing.B) {
	data := []byte(jsonstr)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var i interface{}
		json.Unmarshal(data, &i)
	}
}
