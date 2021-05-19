package jsonschema

import (
	//"context"
	"encoding/json"
	"fmt"
	//"github.com/qri-io/jsonschema"
	"testing"//
)

func TestCreateNew(t *testing.T) {
	var f Schema
	if err := json.Unmarshal(schema, &f); err != nil {
		panic(err)
	}
	iv := map[string]interface{}{
		"a": map[string]interface{}{
			"a1": "",
			"a2": "1",
			"a3": "1",
			"a4": float64(-8),
		},
		"b": map[string]interface{}{
			//"a1": "dd",
			"a2": "1",
			"a3": "1",
			"a4": "",
			"b6": "",
		},
		"c": map[string]interface{}{
			"a1": "",
			"a2": "1",
			"a3": "1",
			"a5": float64(-5.1),
			"a9": float64(0),
		},
		//"age":"4",
		//"fs":3,
		//"sons":[]interface{}{1,2,3},
	}
	type req struct {
		Name string `json:"name"`
		Any  string `json:"any"`
	}
	r := &req{
		Name: "jake2",
	}
	var errs error
	for i := 0; i < 1; i++ {
		//var errs = []Error{}
		errs = f.Validate(iv)
		//errs =f.Validate(r)
		//fmt.Println(errs)

	}
	fmt.Println(r, iv, errs)
	//jsonschema.Properties{}
	//var a  interface{} = 1
	//var b float64 = 1
	//fmt.Println(reflect.DeepEqual(a,b))
}

//func TestCreateNew2(t *testing.T){
//
//	sc:=&jsonschema.Schema{}
//	if err:=json.Unmarshal(schema,sc);err != nil{
//		panic(err)
//	}
//	iv:=map[string]interface{}{
//		"a":map[string]interface{}{
//			"a1":"23",
//			"a2":"1",
//			"a3":"1",
//			"a4":"1",
//		},
//		"b":map[string]interface{}{
//			"a1":"1",
//			"a2":"1",
//			"a3":"1",
//			"a4":"1",
//		},
//		"c":map[string]interface{}{
//			"a1":"1",
//			"a2":"1",
//			"a3":"1",
//			"a4":"5",
//		},
//		//"age":"4",
//		//"fs":3,
//		//"sons":[]interface{}{1,2,3},
//	}
//	for i:=0;i<100000;i++{
//		//var errs = []Error{}
//		sc.Validate(context.Background(),iv)
//		//fmt.Println(errs)
//		//fmt.Println(st.Errs)
//	}
//}

var schema = []byte(`
{

  "type": "object",
  "properties": {
    "a": {
      "switch":"a1",
      "case":{
			"a":{"required":["b1","c1"]},
			"b":{"required":["b2","c2"]}
		},
		"default":{},
      "type": "object",
      "properties": {
        "a1": {
          "type": "string",
          "maxLength": 5
        },
        "a2": {
          "type": "string",
          "maxLength": 5
        },
        "a3": {
          "type": "string",
          "maxLength": 5
        },
        "a4": {"type": "string|number","multipleOf":4}
      }
    },
    "b": {
      "type": "object",
      "if":{
			"required":["a1"]
		},
		"then":{
			"required":["b5"]
		},
		"else":{"required":["b6"]},
      "properties": {
        "a1": {
          "type": "string",
          "maxLength": 5,
          "enum":["dd"]
        },
        "a2": {
          "type": "string"
        },
        "a3": {
          "type": "string",
          "maxLength": 5
        },
        "a4": {
          "type": "string"
        },
        "b6": {
          "type": "string"
        }
      }
    },
    "c": {
      "type": "object",
      "additionalProperties":true,
      "properties": {
        "a1": {
          "type": "string",
          "maxLength": 0
        },
        "a2": {
          "type": "string"
        },
        "a3": {
          "type": "string",
          "maxLength": 5
        },
        "a4": {
          "type": "string"
        },
		"a5":{
			"type":"integer",
			"maximum":0
		}
      }
    }
  }
}
`)

type O2 struct {
	Es string `json:"es" enum:"123,456"`
}

type ObjectTe struct {
	O2
	Name   string   `json:"name" format:"ipv4"`
	Values []string `json:"values" maxLength:"5" enum:"1,2,3,4,5" pattern:"123"`
	Age int `json:"age" minimum:"1" maximum:"100"`
}


func TestNewSchema(t *testing.T) {
	o := &ObjectTe{}
	s ,err  :=GenerateSchema(o)
	if err != nil{
		panic(err)
	}

	err =s.Validate(o)
	fmt.Println(err,string(s.FormatBytes()))
}
