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
	"name":"haha",
	"age":5,
	"sig":"c7cc5f6c2ae8a2bd98189e50872bfd1e",
	"timestamp":5,
	"school":"wh",
	"tss":33,
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

	"set":{
		"userinfo":["append()","${name}",":","${age}"],
		"user_info":["sprintf()","name:%s  age:%v","${name}","${age}"],
		"tm":["dateFormat()","${timestamp}","2006-01-02 15:04:05.999999999 - 0700 MST"],
		"smp":["toJson()","${results}"]
	},
	"and":[
		{"set":{"res":"new()"}},
		{
			"set":{
				"res.code":5,
				"res.message":"success",
				"res.data":{
					"name":"str"
				}
			}
		},
		{
			"if":{
				"neq":{
					"school":"wh"
				}
			},
			"then":{
				"set":{
					"skip_it":true
				}
			},
			"else":{
				"error":["sprintf()","invalid school '%v'","${school}"]
			}
		},
		{
			"if":{
				"required":["results"]
			},
			"then":{
				"setMap":{
					"key":["append()","${name}","${age}"],
					"val":"11"
				}
			}
		},
		{
			"if":{
				"not":{
					"eq":{
						"sig":["md5sum()","${name}","${timestamp}","secret1"]
					}
				}
			},
			"then":{
				"error":"sig not match"
			}
		},
		{
			"if":{
				"not":{
					"lt":{
						"timestamp":["add()","nowtime()",300]
					},
					"gt":{
						"timestamp":["add()","nowtime()",-300]
					}
				}
			},
			"then":{
				"error":"time is valid"
			}
		}
	],
	"additionalProperties":true,
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
			"startWith":"b",
			"maxLength":32,
			"endWith":".json"
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
		t.Error(err)
		return
	}

	var obj any

	err = json.Unmarshal([]byte(exampleJSON), &obj)
	if err != nil {
		t.Error(err)
		return
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
			//panic(err)
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
func BenchmarkIF(b *testing.B) {
	b.ReportAllocs()

	sc, err := NewSchemaFromJSON([]byte(`
{
	"and":[
		{
			"if":{
				"not":{
					"gt":{
						"name":1
					}
				}
				
			},
			"then":{
				
			}
		}
	]
	
}

`))
	if err != nil {
		panic(err)
	}

	var obj any

	err = json.Unmarshal([]byte(`
{
	"name":0
}
`), &obj)
	if err != nil {
		panic(err)
	}
	for i := 0; i < b.N; i++ {
		// TODO: Your Code Here
		err = sc.ValidateObject(obj)
		if err != nil {
			//panic(err)
		}
	}
}
func TestForeach(t *testing.T) {
	sc, err := NewSchemaFromJSON([]byte(`[
{
	"if":{
		"not":{
			"ipIn":{
				"ip":["1.1.1.1"]
			},
			"eq":{
				"hd.username":"1003"
			}
		}
		
	},
	"then":{

		"error":"invalid client ip",
		"set":{
			"ess":"333__${time.now()}_${username}",
			"ess[0]":"333__${time.now()}_${username}"
		}
	}
},
{
	"setExpr":{
		"${username}:${ip}":"true"
	}
},

{
	"foreach":{
		"ws":{
			"foreach":{
				"__val.w":{
					"set":{
						"line":"${line}${__val.c}"
					}
				}
			}
		}
	}
},
{
	"delete":["ws3ss"]
}
]
`))
	if err != nil {
		panic(err)
	}

	var obj any

	err = json.Unmarshal([]byte(`{
"username":"100",
"ip":"1.1.1.1",
"ws":[
	{
		"w":[
			{
				"c":"ni"
			},
			{
				"c":"hao"
			}
		]
	},
	{
		"w":[
			{
				"c":"hello"
			},
			{
				"c":"world"
			}
		]
	}	
]
}
`), &obj)
	if err != nil {
		panic(err)
	}
	err = sc.Validate(obj)
	fmt.Println(err)

	res, _ := json.MarshalIndent(obj, "", "\t")
	fmt.Println(string(res))
}

//

func BenchmarkFOR(b *testing.B) {
	b.ReportAllocs()
	fmt.Println(b.N)
	sc, err := NewSchemaFromJSON([]byte(`[
{
	"set":{
		"an,a":"1",
		"age":2,
		"ce":3
	},
	"setExpr":{
		"${age}_${ce}":"true"
	}
}
]
`))
	if err != nil {
		panic(err)
	}

	var obj any

	err = json.Unmarshal([]byte(`{
"ws":[
	{
		"w":[
			{
				"c":"ni"
			},
			{
				"c":"hao"
			}
		]
	},
	{
		"w":[
			{
				"c":"hello"
			},
			{
				"c":"world"
			}
		]
	}	
]
}
`), &obj)
	if err != nil {
		panic(err)
	}
	err = sc.Validate(obj)
	fmt.Println(err)

	res, _ := json.MarshalIndent(obj, "", "\t")
	fmt.Println(string(res))

	//f, err := os.Create("cpu.perf")
	//if err != nil {
	//	panic(err)
	//}
	//pprof.StartCPUProfile(f)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = sc.ValidateObject(obj)
		if err != nil {
		}
	}
	//pprof.StopCPUProfile()
}

func BenchmarkJP(b *testing.B) {

	b.ReportAllocs()

	jp, err := parseJpathCompiled("name")
	if err != nil {
		panic(err)
	}
	mm := map[string]any{
		"name": 5,
	}
	for i := 0; i < b.N; i++ {
		jp.Set(mm, 1)
	}

}

// 6nm
func TestParse(t *testing.T) {

}
