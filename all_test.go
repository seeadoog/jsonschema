package jsonschema

import (
	"encoding/json"
	"fmt"
	"github.com/yuin/gopher-lua"
	"net/http"
	"reflect"
	"testing"
	"time"
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

func TestLua(t *testing.T) {
	lvm := lua.NewState()

	lf ,err := lvm.LoadString(`
if get_name() == 't' then
	set_val('af','b')
end
if not a then a = 0 else a = a +1 end 

set_val('a',a)
`)
	if err != nil{
		panic(err)
	}
	//lvm = lua.NewState()
	lvm.SetGlobal("get_name",lvm.NewFunction(func(l *lua.LState) int {
		l.Push(lua.LString("t"))
		return 1
	}))

	m := map[string]interface{}{}
	lvm.SetGlobal("set_val",lvm.NewFunction(func(l *lua.LState) int {
		key := l.ToString(1)
		val:= l.ToString(2)
		m[key] = val
		return 0
	}))

	for i := 0; i < 100000; i++ {
		lvm.Push(lf)
		err := lvm.PCall(0, lua.MultRet, nil)
		if err != nil{
			panic(err)
		}
	}
	fmt.Println(m)
}


func TestLua2(t *testing.T) {
	lvm := lua.NewState()

	for range [5]bool{}{

	}
	//lvm = lua.NewState()
	lvm.SetGlobal("get_name",lvm.NewFunction(func(l *lua.LState) int {
		l.Push(lua.LString("t"))
		return 1
	}))

	m := map[string]interface{}{}
	lvm.SetGlobal("set_val",lvm.NewFunction(func(l *lua.LState) int {
		key := l.ToString(1)
		val:= l.ToString(2)
		m[key] = val
		return 0
	}))

	for i := 0; i < 100000; i++ {
		lvm.DoString(`
if get_name() == 't' then
	set_val('a','b')
end
`)
	}
	fmt.Println(m)
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
	err := json.Unmarshal([]byte(schema),sc)
	if err != nil{
		panic(err)
	}
	obj := &Object{
		Ids: []Ids{{Name: int(5)}},
	}


	err = sc.Validate(obj)
	if err != nil{
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
	validate(data,`{"param":"50v","g":"50"}`)


}

type messgae struct {
	data string
	writer func(msg string)
}

func TestLUa3(t *testing.T) {
	l := lua.NewState()
	//ctx, _ := context.WithTimeout(context.Background(),5000*time.Millisecond)
	//l.SetContext(ctx)
	c := make(chan messgae,10)
	go func() {
		http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			done := make(chan int)
			c <- messgae{
				data: string(request.RequestURI),
				writer: func(msg string) {
					writer.Write([]byte(msg))
					done <- 1
				},
			}
			<- done
		})
		panic(http.ListenAndServe(":8762",nil))
	}()
	l.SetGlobal("get_message",l.NewFunction(func(l *lua.LState) int {
		data := <- c
		tb := &lua.LTable{}
		tb.RawSetString("write_response",l.NewFunction(func(state *lua.LState) int {
			msg := state.ToString(1)
			data.writer(msg)
			return 0
		}))
		tb.RawSetString("data",lua.LString(data.data))
		l.Push(tb)
		return 1
	}))

	l.SetGlobal("sleep",l.NewFunction(func(state *lua.LState) int {
		fmt.Println("sloop",state.ToInt(1))
		time.Sleep(time.Duration(state.ToInt(1))*time.Millisecond)
		fmt.Println("sloop2")
		return 0
	}))

	err := l.DoString(`
function request(msg)
	sleep(4000)
	print('msg',msg.data)
	
	msg.write_response('tounima'..msg.data)
end
`)

//	go func() {
//		time.Sleep(1*time.Second)
//		for i := 0; i < 1000; i++ {
//			func() {
//				err = l.DoString(`
//function request(msg)
//	print('msg',msg.data)
//	msg.write_response('tounimaddd'..msg.data)
//end
//`)
//			}()
//		}
//	}()

	if err != nil{
		panic(err)
	}

	err = l.DoString(`
function do_request(msg)
	print('handler request')
	local res = request(msg)
end
`)
	if err != nil{
		panic(err)
	}

	err = l.DoString(`
	for i=1,5 do
		local coo = coroutine.create(
		function(msg)
			sleep(4000)
			print('sleep')
		end
	
	)
	coroutine.resume(coo, msg)
	end


`)
	if err != nil{
		panic(err)
	}

	err = l.DoString(`
local c = 0
while(true)
do 
	local msg = get_message()
	local coo = coroutine.create(
		function(msg) 
			do_request(msg)
		end
	
	)
	coroutine.resume(coo, msg)
	-- coroutine.resume(coo, msg)
end

`)
	fmt.Println(err)
	select {

	}
}

/*
   100000
   80000
 */
