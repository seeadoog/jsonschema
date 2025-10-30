package jsonschema

import (
	"encoding/json"
	"fmt"
	"github.com/seeadoog/jsonschema/v2/utils"
	"testing"
)

type User struct {
	Name   *string                 `json:"name" enums:"1,2,3,4,56" maxLength:"5"`
	Age    *int                    `json:"age"`
	Sister *map[string]interface{} `json:"sister"`
	Childs [2]*User                `json:"childs"`
	Msg    []byte                  `json:"msg"`
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
		"msg": ([]byte("hello world")),
	}

	var v User
	err := UnmarshalFromMap(m, &v)
	if err != nil {
		panic(err)
	}
	fmt.Println(v)
	fmt.Println(*v.Sister)
	fmt.Println(string(v.Msg))
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

type B struct {
	Birth int `json:"birth"`
}
type A struct {
	B
	Name string `json:"name,omitempty" maxLength:"14" format:"phone"`
	Age  *int   `json:"age" maximum:"100" minimum:"0" enum:"1,2,3,4,5" format:"ip"`
	Addr []int  `json:"ace"`
}

func TestDecode(t *testing.T) {

	sc, err := GenerateSchema(&A{})
	if err != nil {
		panic(err)
	}

	bb, _ := sc.MarshalJSON()
	fmt.Println(string(bb))
	a := &A{}
	err = sc.ValidateAndUnmarshalJSON([]byte(`
{
	"name":"ddf",
	"age":50,
	"birth":5,
	"ace":4
}
`), a)
	if err != nil {
		t.Error(err)
		//return
	}

	fmt.Println(a)
}

//	func TestNewSchema2(t *testing.T) {
//		sc ,err:= NewSchemaFromJSON([]byte(`
//
//	{
//		"type":"object",
//		"properties":{
//			"name":{
//				"type":"string"
//			}
//		}
//	}
//
// `))
//
//		if err != nil{
//			panic(err)
//		},
//	}

func TestIndexRange(t *testing.T) {
	utils.IndexRange("a,b,c,d,e,f", ',', func(idx int, s string) bool {
		fmt.Println(idx, s)
		return true
	})

	utils.IndexRange("abc", ',', func(idx int, s string) bool {
		fmt.Println(idx, s)
		return true
	})
	utils.IndexRange("", ',', func(idx int, s string) bool {
		fmt.Println(idx, s)
		return true
	})

}

func Test_Switch(t *testing.T) {
	sc := `
	{
		"type":"object",
		"switch":"method",
		"case":{
			"get":{
					"properties":{
						"method":{
							"type":"string"
						},
						"get_1":{
							"type":"string"
						}
					}   
			},
			"post":{
					"properties":{
						"method":{
							"type":"string"
						},
						"post_1":{
							"type":"string"
						},
						 "post_2":{
							"type":"string"
						}
					}
			}
		},
		"defaults":{
			"error":["sprintf()","not support method '%v'","${method}"]
		},
		"if":{
			"eq":{
				"method":"get"
			},
			"startWiths":{
				"path":"/user"
			}
		},
		"then":{
			"error":["sprintf()","path is not allowed to access '%v'","${path}"]
		},
		"else":{
			"set":{
				"exit_result":{
					"code":0,
					"message":"pass the pass",
					"infos":{
						"method":"${method}"
					}
				}
			}
		}	
	}
	
	`
	m := &Schema{}

	err := json.Unmarshal([]byte(sc), m)
	if err != nil {
		panic(err)
	}
	req := map[string]any{
		"method": "get",
		"path":   "/use/login",
	}
	fmt.Println(m.Validate(req))
	fmt.Println(req)

}
