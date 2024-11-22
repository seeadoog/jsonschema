package jsonschema

import (
	//"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/xeipuuv/gojsonschema"

	"github.com/xeipuuv/gojsonpointer"
	"testing" //
)

func TestJ(t *testing.T) {
	p, err := gojsonpointer.NewJsonPointer("/a")
	if err != nil {
		t.Fatal(err)
	}
	v := map[string]interface{}{
		"aaa": 2,
	}
	_, err = p.Set(v, 1)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(v)
}

func BenchmarkJPs(t *testing.B) {
	p, err := gojsonpointer.NewJsonPointer("/aa")
	if err != nil {
		t.Fatal(err)
	}
	v := map[string]interface{}{
		"aaa": 2,
	}
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		p.Set(v, 1)
	}

}

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
	Es string `json:"es" enum:"123,456" required:"true"`
}

type ObjectTe struct {
	O2
	Name   string   `json:"name" format:"ipv4" test:"3" required:"true"`
	Values []string `json:"values" maxLength:"5" enum:"1,2,3,4,5" pattern:"123"`
	Age    int      `json:"age" minimum:"1" maximum:"100"`
	O3     *O2      `json:"o3" required:"true"`
	DnS    *float64 `json:"dn_s" minimum:"1.1" maximum:"2.2"`
}

type test int

func (t test) Validate(c *ValidateCtx, value interface{}) {

}

var newTest NewValidatorFunc = func(i interface{}, path string, parent Validator) (Validator, error) {

	return new(test), nil
}

func float(v float64) *float64 {
	return &v
}

func TestNewSchema(t *testing.T) {
	RegisterValidator("test", newTest)
	AddRefString("test")
	a := "1.1.1.1"
	o := &ObjectTe{
		Name: a,
		O3: &O2{
			Es: "123",
		},
		Age: 100,
		DnS: float(1.4),
	}
	s, err := GenerateSchema(o)
	if err != nil {
		panic(err)
	}

	err = s.Validate(o)
	fmt.Println(err, string(s.FormatBytes()))
}

func BenchmarkSchema(b *testing.B) {
	sc := &Schema{}

	err := json.Unmarshal([]byte(`
{
	"type":"object",
	"properties":{
		"name":{
			"type":"string",
			"maxLength":50,
			"enum":["s","b"],
			"pattern":"^[sb]{1,2}$"
		},
		"age":{
			"type":"integer",
			"maximum":50
		},
		"birthday":{
			"type":"string"
		},
		"cs":{
				"type":"object",
				"properties":{
					"name":{
						"type":"string"
					}
				}
			}
	}
}


`), sc)
	if err != nil {
		panic(err)
	}
	o := objOfJson(`
{
	"name":"s",
	"age":5,
	"birthday":"20211102",
	"cs":{
		"name":"4"
	}
}

`)

	err = sc.Validate(o)
	if err != nil {
		panic(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sc.Validate(o)
	}
}

func BenchmarkSchema2(b *testing.B) {

	o := objOfJson(`
{
	"name":33,
	"age":"ddd",
	"birthday":"20211102",
	"cs":{
		"name":"dd",
		"age":5
	}
}

`)
	loader := gojsonschema.NewBytesLoader([]byte(`
{
	"type":"object",

	"properties":{
		"name":{
			"type":"string",
			"maxLength":50,
			"enum":["s","b"]
		},
		"age":{
			"type":"integer",
			"maximum":50
		},
		"birthday":{
			"type":"string"
		},
		"cs":{
			"type":"object",
			"properties":{
				"name":{
					"type":"string"
				},
			  "age":{
					"type":"integer",
					"maximum":50
				}
			}
		}
	}
}


`))
	sc, err := gojsonschema.NewSchema(loader)
	if err != nil {
		panic(err)
	}
	ooo := gojsonschema.NewRawLoader(o)
	re, err := sc.Validate(ooo)
	if err != nil {
		panic(err)
	}
	if re != nil {
		fmt.Println("errors:", re.Errors())

	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sc.Validate(ooo)
	}
}

func objOfJson(in string) interface{} {
	var i interface{}
	err := json.Unmarshal([]byte(in), &i)
	if err != nil {
		panic(err)
	}
	return i
}

type Cars struct {
	ans         int // 答案
	firstChose  int
	sencodChose int
	open        int
}

func (c *Cars) openWindow() {
	for i := 0; i < 3; i++ {
		if i != c.ans && i != c.firstChose {
			c.open = i
			return
		}
	}
}

func (c *Cars) choseFirst() {
	c.firstChose = rand.Int() % 3
}

func (c *Cars) switchWindow() {

	for i := 0; i < 3; i++ {
		if i != c.firstChose && i != c.open {
			c.sencodChose = i
		}
	}
}

func (c *Cars) notSwitch() {
	c.sencodChose = c.firstChose
}
func (c *Cars) getCar() bool {
	return c.sencodChose == c.ans
}
func (c *Cars) runWithCHose() bool {
	c.choseFirst()
	c.openWindow()
	c.switchWindow()
	return c.getCar()
}

func (c *Cars) runWithoutCHose() bool {
	c.choseFirst()
	c.openWindow()
	c.notSwitch()
	return c.getCar()
}

func (c *Cars) runRand() bool {
	if rand.Int()%2 == 0 {
		c.notSwitch()
	} else {
		c.switchWindow()
	}
	return c.getCar()
}

func TestCars(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	wins := 0
	for i := 0; i < 100000; i++ {
		c := &Cars{
			ans: 0,
		}
		if c.runWithCHose() {
			wins++
		}
	}
	fmt.Println(wins)
}

func TestCarsWithoutCHose(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	wins := 0
	for i := 0; i < 100000; i++ {
		c := &Cars{
			ans: 0,
		}
		if c.runWithoutCHose() {
			wins++
		}
	}
	fmt.Println(wins)
}

func TestCarsWithRan(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	wins := 0
	for i := 0; i < 100000; i++ {
		c := &Cars{
			ans: 0,
		}
		if c.runRand() {
			wins++
		}
	}
	fmt.Println(wins)
}

func TestDefault(t *testing.T) {
	type User struct {
		Name   string   `json:"name" maxLength:"15" pattern:"^[0-9a-zA-Z_\\-.]+$"`
		Age    int      `json:"age" maximum:"150" minimum:"1" multipleOf:"2"`
		Childs []string `json:"childs" uniqueItems:"true" maxItems:"5" minItems:"2"`
	}

	sc, err := GenerateSchema(&User{})
	if err != nil {
		panic(err)
	}

	fmt.Println(string(sc.Bytes()))

}

//{}

func TestRef(t *testing.T) {
	sc := `

{
  "$defs": {
    "user": {
      "type": "object",
      "properties": {
        "name": {
          "type": "string"
        },
        "age": {
          "type": "integer"
        },
        "child": {
          "$ref": "#/"
        },
        "sams": {
          "properties": {
            "gcc": {
              "type": "string"
            },
            "scc": {
              "$ref": "#/$defs/user/properties/sams"
            }
          }
        }
      }
    }
  },
  "$ref": "#/$defs/user"
}



`
	//$.child
	//$.sams.scc
	ss, err := NewSchemaFromJSON([]byte(sc))
	if err != nil {
		panic(err)
	}
	err = ss.Validate(map[string]any{
		"sams": map[string]any{
			"gcc": "",
		},
		"child": map[string]any{
			"sams": map[string]any{
				"gcc": "",
				"scc": map[string]any{
					"gcc": "",
				},
			},
			"child": map[string]any{
				"sams": map[string]any{
					"gcc": "",
					"scc": map[string]any{
						"gcc": "",
					},
				},
				"child": map[string]any{
					"sams": map[string]any{},
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}

}

func TestDefaultInner(t *testing.T) {
	sc := `
{
	"type":"object",
    "defaultVals":{
		"name":"xxxxx",
		"smdts":"chenjian"
	},
	"properties":{
		"name":{
			"type":"string"
		},
		"age":{
			"type":"integer",
			"exclusiveMaximum":15 ,
			"maximum":15,
			"minimum":3,
			"exclusiveMinimum":false
		},
		"child":{
			"properties":{
				"name":{
					"defaultVal":"xx"
				}
			},
			"defaultVal":{}
		}
	}
	
}

`
	ss, err := NewSchemaFromJSON([]byte(sc))
	if err != nil {
		panic(err)
	}

	c := map[string]any{
		"name": "ddddd",
		"age":  float64(3),
	}
	err = ss.Validate(c)
	if err != nil {
		panic(err)
	}

	fmt.Println(c)

}

//

func Test_SSchema(t *testing.T) {
	f, err := ioutil.ReadFile("/Users/sjliu/temp/schemadraft")
	if err != nil {
		t.Error(err)
		return
	}
	f = []byte(`
	{
		"$defs":{
			"arrayInt":{
				"type":"array",
				"items":{
					"type":"integer"
				}
			}
		},
		"properties":{
			"type":{
				"type":"string",
				"enums":["string"]
			},
			"properties":{
				"additionalProperties":{
					"$ref":"#/$defs/arrayInt"
				},
				"properties":{}
			}
		}
	}

	`)
	sc := &Schema{}

	err = json.Unmarshal(f, sc)
	if err != nil {
		panic(err)
	}

	err = sc.Validate([]byte(`
	{
		"type":"string",
		"properties":{
			"ancds":{
				"type":"string2"
			}
		}
	}
	
	`))

	fmt.Println(err)

	// $,properties[]
}
func TestSrr(t *testing.T) {
	ss := "你说地方"
	fmt.Println(ss[:4])
}

func TestParseAsss(t *testing.T) {
	vs, err := parseComboValue("aseln.${line}_${__val\\.c}")
	if err != nil {
		panic(err)
	}
	//fmt.Println(len(vs.values))
	fmt.Println(vs.Get(map[string]any{
		"line":    "name",
		"__val.c": "age",
	}))
	parseValue("")

	fmt.Println(eq("1", []string{}))

	fmt.Println(eq("", nil))

}

//
