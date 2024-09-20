package jsonschema

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	qsc "github.com/qri-io/jsonschema"
	"github.com/xeipuuv/gojsonschema"
)

var exampleJSON = `
{
	"name":"xiaohu",
	"age":5,
	"school":"wangchen",
	"hobby":["ball","game","music"],
	"results":{
		"code":100,
		"message":"success",
		"data":{
			"method":"GET",
			"text":"hello",
			"desc":{
				"encoding":"raw",
				"format":"160"
			}
		}
	},
	"results2":{
		"code":100,
		"message":"success",
		"data":{
			"method":"GET",
			"text":"hello",
			"desc":{
				"encoding":"raw",
				"format":"160"
			}
		}
	}
}

`

var exampleSchema = `
{
    "if":{
		"required":["name"]
	},
	"then":{
		"set":{
			"name":{
				"func":"append",
				"args":[
					"${name}",":",
					{
						"func":"join",
						"args":["${hobby}",","]
					}
				]
			}
		}
	},
	"properties": {
		"to_sc":{},
		"age": {
			"type": "number",
			"maximum":100,
			"minimum":0
		},
		"hobby": {
			"items": {
				"type": "string",
				"enum":["ball","game","music"]
			},
			"type": "array"
		},
		"name": {
			"type": "string",
			"maxLength":32
		},
		"results": {
			"properties": {
				"code": {
					"type": "number"
				},
				"data": {
					"properties": {
						"desc": {
							"properties": {
								"encoding": {
									"type": "string"
								},
								"format": {
									"type": "string"
								}
							},
							"type": "object"
						},
						"method": {
							"type": "string"
						},
						"text": {
							"type": "string"
						}
					},
					"type": "object"
				},
				"message": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"results2": {
			"properties": {
				"code": {
					"type": "number"
				},
				"data": {
					"properties": {
						"desc": {
							"properties": {
								"encoding": {
									"type": "string"
								},
								"format": {
									"type": "string"
								}
							},
							"type": "object"
						},
						"method": {
							"type": "string"
						},
						"text": {
							"type": "string"
						}
					},
					"type": "object"
				},
				"message": {
					"type": "string"
				}
			},
			"type": "object"
		},
		"school": {
			"type": "string"
		}
	},
	"type": "object"
}


`

func genSchemaFromJSON(in string) string {
	var i any
	json.Unmarshal([]byte(in), &i)
	res := map[string]any{}

	genSchema(i, res)

	bs, _ := json.Marshal(res)
	return string(bs)
}

func genSchema(t any, sc map[string]any) {
	switch v := t.(type) {
	case string:
		sc["type"] = "string"
	case map[string]any:
		sc["type"] = "object"
		props := map[string]any{}
		sc["properties"] = props
		for k, v := range v {
			pt := map[string]any{}
			genSchema(v, pt)
			props[k] = pt
		}
	case float64:
		sc["type"] = "number"
	case bool:
		sc["type"] = "boolean"
	case []any:
		sc["type"] = "array"
		if len(v) > 0 {
			items := map[string]any{}
			sc["items"] = items
			genSchema(v[0], items)
		}
	case nil:

	default:
		panic("invalid type:" + reflect.TypeOf(t).String())
	}
}

func Test_JSONSC(t *testing.T) {
	fmt.Println(genSchemaFromJSON(exampleJSON))
}

func TestNewSchemaFromJSON(t *testing.T) {
	sc, err := NewSchemaFromJSON([]byte(exampleSchema))
	if err != nil {
		panic(err)
	}

	var obj any

	err = json.Unmarshal([]byte(exampleJSON), &obj)
	if err != nil {
		panic(err)
	}
	err = sc.Validate(obj)
	fmt.Println(err)

	res, _ := json.MarshalIndent(obj, "", "\t")
	fmt.Println(string(res))
}
func BenchmarkSchema_local(b *testing.B) {
	// TODO: Initialize
	b.ReportAllocs()

	sc, err := NewSchemaFromJSON([]byte(exampleSchema))
	if err != nil {
		panic(err)
	}

	var obj any

	err = json.Unmarshal([]byte(exampleJSON), &obj)
	if err != nil {
		panic(err)
	}
	for i := 0; i < b.N; i++ {
		// TODO: Your Code Here
		err = sc.ValidateObject(obj)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkSchema_gojsonschema(b *testing.B) {
	// TODO: Initialize
	b.ReportAllocs()

	loader := gojsonschema.NewBytesLoader([]byte(exampleSchema))

	sc, err := gojsonschema.NewSchema(loader)
	if err != nil {
		panic(err)
	}
	ooo := gojsonschema.NewBytesLoader([]byte(exampleJSON))

	// if re != nil {
	// 	fmt.Println("errors:", re.Errors())

	// }

	for i := 0; i < b.N; i++ {
		// TODO: Your Code Here
		_, err := sc.Validate(ooo)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkSchema_qri_io_jsonschema(b *testing.B) {
	b.ReportAllocs()
	// TODO: Initialize
	var sc = qsc.Must(exampleSchema)
	var obj any

	err := json.Unmarshal([]byte(exampleJSON), &obj)
	if err != nil {
		panic(err)
	}

	for i := 0; i < b.N; i++ {
		// TODO: Your Code Here
		sc.Validate(context.Background(), obj)
	}
}
