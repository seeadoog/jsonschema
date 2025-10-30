package expr

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestFuncJudge(t *testing.T) {
	fmt.Println()
}

func Test_isFuncCall(t *testing.T) {
	type args struct {
		e string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{args: args{e: "f"}, want: false},
		{args: args{e: "()"}, want: false},
		{args: args{e: "f()"}, want: true},
		{args: args{e: "pl(k,v)"}, want: true},
		{args: args{e: "f(a,b)"}, want: true},
		{args: args{e: "f(a)"}, want: true},
		{args: args{e: "$(a)"}, want: true},
		{args: args{e: "f(a,f(b),b)"}, want: true},
		{args: args{e: "f.ab(a,f(b),b)"}, want: true},
		{args: args{e: "a=ff(a,f(b),b)"}, want: false},
		{args: args{e: "name=a"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isFuncCall(tt.args.e); got != tt.want {
				t.Errorf("isFuncCall() = %v, want %v arg:%v", got, tt.want, tt.args.e)
			}
		})
	}
}

func Test_isSetCond(t *testing.T) {
	type args struct {
		e string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{args: args{e: "f=b"}, want: true},
		{args: args{e: "set-es=b"}, want: true},
		{args: args{e: "ad.set-es.000=b"}, want: true},
		{args: args{e: "f='ff'"}, want: true},
		{args: args{e: "$.f.b[0]='ff'"}, want: true},
		{args: args{e: "f=b()"}, want: true},
		{args: args{e: "f=b(a)"}, want: true},
		{args: args{e: "f=b(a,b)"}, want: true},
		{args: args{e: "f=b(a,b,c())"}, want: true},
		{args: args{e: "f=b(a,b,c(d,'d'))"}, want: true},
		{args: args{e: "f"}, want: false},
		{args: args{e: "f()"}, want: false},
		{args: args{e: "a == b"}, want: false},
		{args: args{e: "a==b"}, want: false},
		{args: args{e: "a=b"}, want: true},
		{args: args{e: "a>=b"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSetCond(tt.args.e); got != tt.want {
				t.Errorf("isSetCond() = %v, want %v", got, tt.want)
			}
		})
	}
}

func rawval(m map[string]any) any {
	if m["a"] == nil {
		m["a"] = make(map[string]any)
	}
	if m["bss"] != nil {
		m["a"].(map[string]any)["bs"] = m["bs"]
	} else {
		m["a"].(map[string]any)["bs"] = time.Now().Format("2006-01-02 15:04:05")
	}
	return m["a"]
}
func BenchmarkRaw(b *testing.B) {
	var a map[string]any = make(map[string]any)
	for i := 0; i < b.N; i++ {
		rawval(a)
	}
}
func BenchmarkExpr(b *testing.B) {
	e, err := parseExpr("name=3")
	if err != nil {
		panic(err)
	}
	b.ReportAllocs()
	RegisterDynamicFunc("test", 2)
	ctx := &Context{
		table: map[string]any{
			"bs": []any{"1", "2", "3", "4"},
		},
	}
	err = e.Exec(ctx)
	for i := 0; i < b.N; i++ {
		e.Exec(ctx)
	}
}

func BenchmarkMapSet(b *testing.B) {
	m := make(map[string]any)
	for i := 0; i < b.N; i++ {
		m["name"] = 3
	}
}

func TestForReg(t *testing.T) {
	// a=cmd?abc?c:d:e=f
	fmt.Println(forRegexp.FindAllStringSubmatch("k , v in abc", -1))
}

var (
	scpt = `
[
"ab.name=bss or time_format(time_now(),'2006-01-02 15:04:05')",
"ab.age = bss.name ? abc : ced",
"d= ac or a? 1:2",
"abrr=slice_init(1+3*d,3/2,4,5,'6',slice_init(4,5,6))",
{
	"if":"eqs(ass,'')",
	"then":["ac=abcd or '' "]
},
{
	"for":"k,v in $.pss",
	"do":[]
},
{
	"switch":"$.name",
	"case":{
		"'perter'":"$.route='/vnt'",
		"'json'":"$.route='/mlo'"
	},
	"default":"print('goto default')"
},
"print(type(arr))",

"mm.name='5ds3'",
"if(eq(mm.name,'5ds'),$.a==5)",
"fff.ss=5",
"mm.age=number(3)",
"print('nameis:',mm.name)",
"ss[0]=1",
"ss[1]=2",
"ss=append(ss,3,4)",
"sss=slice_cut(ss,0,1)",
"sssa=str_join(ss,'|')",
"ss='asdnds'",
"ss3=ternary(mm.name,'true','false')",
"s5=hex_encode(md5_sum(ss))",
"auths=sprintf('host: %v \ndate: %v','app.xxc.om',time_now())",
"sha=hex_encode(hmac_sha256(auths,'helloworld'))",
"header.X-Http-Name='ems\\''",
"#hres=http_request('GET','http://172.30.209.27',nil,nil,10)",
"gs=json_from(hres.body)",
"msg=gs.message",
"ges=d.d.d",
"a=5",
"bd=a==5",
"ub.gte=a>=5",
"ub.lte=a<=5",
"ub.gt=a>3",
"ub.lt=a<1",
"ce=a!=3",
"c5=(a+3+4)*5",
"ub.pow=3^3+6",
{
	"switch":{
		"mm.name=='5ds3'":"def.aa=1",
		"lt(mm.age,2)":"def.aa=2"
	},
	"default":["def.aa=5"]
},
"return(3)"
]
`
)

func BenchmarkParse(b *testing.B) {
	var o any
	err := json.Unmarshal([]byte(scpt), &o)
	if err != nil {
		panic(err)
	}
	for i := 0; i < b.N; i++ {
		ParseFromJSONStr(scpt)
	}
}

func benchGORaw(table map[string]any) {
	table["name"] = "jhon"
	table["age"] = float64(23)
	if table["age"].(float64) > 20 {
		table["gender"] = "cn"
	} else {
		table["gender"] = "sn"
	}
	if len(table["name"].(string)) > 3 {
		table["name_msg"] = "name too large"
	} else {
		table["name_msg"] = "name is ok"
	}

}
func benchGORaw2(table map[string]any) {
	root := table["$"].(map[string]any)
	if root["name"] == "500" && root["age"].(float64) > 30 {
		root["route"] = "/abc/def"
	} else {
		root["route"] = "/default"
	}

}

func BenchmarkExec(b *testing.B) {
	scpt := `
"time_now()::format('2006-01-02 15:04:05')"
`
	s, err := ParseFromJSONStr(scpt)
	if err != nil {
		panic(err)
	}
	b.ReportAllocs()

	tb := NewContext(map[string]any{
		"$": map[string]any{
			"name": "500",
			"age":  float64(44),
		},
	})

	for i := 0; i < b.N; i++ {
		err := tb.Exec(s)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println(tb)
}

func BenchmarkParseExp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseFromJSONObj(`$.route = $.name == '500' && $.age > 30 ? '/abc/def' : '/default'`)
	}
}

func BenchmarkGORaw(b *testing.B) {

	tb := map[string]any{
		"$": map[string]any{
			"name": "500",
			"age":  float64(44),
		},
	}
	for i := 0; i < b.N; i++ {
		benchGORaw2(tb)
	}
	fmt.Println(tb)
}

func TestJSONScpt(t *testing.T) {
	RegisterDynamicFunc("add2", 2)
	RegisterDynamicFunc("response_write", 1)
	e, err := ParseFromJSONStr(scpt)
	if err != nil {
		panic(err)
	}
	ctx := &Context{
		table: map[string]any{
			"mm": map[string]any{
				"name": "5ds3",
			},

			"bsss": "6",
			//"ass":  "",
			"arr": []any{1, 2, 3, 4},
			"sw":  float64(11),
			"$": map[string]interface{}{
				"name": "perter",
				"pss":  []any{1, 2, 3, 4, 5},
				"ws": []any{
					map[string]any{
						"cw": []any{
							map[string]any{
								"w": "xx",
							},
							map[string]any{
								"w": "bb",
							},
						},
					},
				},
			},
		},
	}
	ctx.SetFunc("add2", FuncDefine2(func(a, b float64) float64 {
		return a + b
	}))

	err = ctx.Exec(e)

	if err != nil {
		_, ok := err.(*Return)
		if !ok {
			panic(err)
		}
	}

}

func vv(v any) string {
	return reflect.ValueOf(v).String()
}

func BenchmarkName(b *testing.B) {
	var a string
	for i := 0; i < b.N; i++ {
		a = vv("x")
	} // a+b+c*d * (d+5)
	fmt.Println(a)
}

var priority = map[byte]int{}

func toSufix(s string) string {
	ss := stack[byte]{}
	token := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '+', '-', '*', '/':
			if ss.empty() {
				ss.push(c)
				continue
			}
			for !ss.empty() && (ss.top() == '+' || ss.top() == '-' || ss.top() == '*' || ss.top() == '/') {
				if priority[ss.top()] >= priority[c] {
					token = append(token, ss.pop())
				} else {
					break
				}
			}
			ss.push(c)

		case '(':
			ss.push(c)
		case ')':
			for {
				cc := ss.pop()
				if cc == '(' {
					break
				}
				token = append(token, cc)
			}
		default:
			token = append(token, c)
		}
	}
	for !ss.empty() {
		token = append(token, ss.pop())
	}
	return string(token)
}

func calcSuffix(s string) int {
	fmt.Println(s)
	ss := stack[int]{}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '+':
			ss.push(ss.pop() + ss.pop())
		case '-':
			a := ss.pop()
			b := ss.pop()
			ss.push(b - a)
		case '*':
			ss.push(ss.pop() * ss.pop())
		case '/':
			a := ss.pop()
			b := ss.pop()
			ss.push(b / a)
		default:

			n, _ := strconv.Atoi(string(c))
			ss.push(n)
		}
	}
	return ss.pop()
}

// ([$.-_0-9a-zA-Z]\(([$.-_0-9a-zA-Z]+)|()|'(.+)'\)^)
func TestToSufix(t *testing.T) {
	fmt.Println(calcSuffix(toSufix("3+4+2*2/(2-1)*5+5")))
}

// abc/-

func TestHTTP(t *testing.T) {
	scpt := `


[
  "name='hello'",
  "age = 5",
  "gender = age > 3 ? 'a1' : 'a2'",
  "gender2 = age > 6 ? 'a1' : 'a2'",
  {
    "if": "age > 3",
    "then": "ageset=1",
    "else": "ageset=2"
  },
  {
    "if": "age > 6",
    "then": "ageset2=1",
    "else": "ageset2=2"
  },
	"$ = new()",
  {
    "for": "k,v in kv",
    "do": [
      "#helloworld",
      "set($,k,v)"
    ]
  },
 "header.name = 'nn'",
  "cbool = age == 5",
  "cb2 = name or 'yes'",
  "cb3 = name2 or 'yes'",
  {
     "switch":"name",
	 "case":{
		"'hello'":"sname='1'"
     }
  },
  {
     "switch":"name",
	 "case":{

		"'hello2'":"sname='1'"
     },
	 "default":"dname='2'"
  },
  {
		"switch":{
			"name=='hello'" :"ssname='12'"
		}
  },
  "fmt = '${name}_${age}' ",
  "fmt2 = '\\\\${name}_${age}' ",
  "fmt3 = 'he\nhe'",
  "fmt5 = name == 'hello' ? 'is_hello' : (age > 1 ? 'age_1':'age_2')",
  "fmt6 = name == 'hello2' ? 'is_hello' : (age > 10 ? 'age_1':'age_2')",
  "$.channel = 'cbc' ",
  "$.calc_func = $.channel == 'vms' ? '${$.channel}.tokens.total' : ( $.channel == 'cbc' ? '${$.channel}.business.total' : 'business.total' ) ",
  "_ = $.channel == 'cbc' ? $.cset = '1'  :  $.cset = '2' ",
  "e = d = f = g = 4",
  "$.channel == 'cbc' ? ($.cset1 = '1') : ($.cset1 = '2')",
  "$.top2 = 4",
  "$.top = $.top? $.top : 6",
  "$.top2 = $.top2? $.top2 : 6",
  "name == '500' && age ==20 ? ($.route='/abc') : ($.route = 'default')",
  "name == 'hello' && age ==5 ? ($.route2='/abc') : ($.route2 = 'default')",
  "age > 3 ? ($.route3 = '/r3'): _",
  "age != 5 ? ($.route4 = '/r3'): ($.route4 = '/r5')",
  "age != 6 ? ($.route5 = '/r3'): ($.route5 = '/r5')",
  "haxp = str_has_prefix(name,'he')",
  "haxpf = str_has_prefix(name,'ge')",
   "cfa = kv['kv.a']",
   "cf2 = kv::c::d",
  "sb = str_builder()::write('hello')::write('world')::string()",
  "haxpn = 'hello world'::has_prefix('hello')",
  "haxpnn = !'hello world'::has_prefix('hello')",
  "sss = slice_init(1,3,4,5,5)",
   "slice_new(3)",
  "ssct = sss::slice(0,2)",
   "b64 = name::base64()::base64d()::string()",
   "a==b && c==d",
	"nilt = abcs::type()",
  "name=='hello'? m1 = 1 ; m2 = 2 ; m3=3 : _",
 {
          "for": "k,v in $2.data",
          "do": "v.id in ['aa','bb']? v.status=v.data.status ; $2::set(v.id,v.data) : (!$2.status? $2.status=v.status;$2.result=v.data : _ )  "
   },
   "header2 = new()::set('content-type', 'text/csv')",
	"now = time_now(); date = '${now::year()}-${now::month()}-${now::day()}' ",
   "mmaps = {'name':'5','age':6,'bdy':{'xm':3},name:name,name::type() : name::type(),'xx1': age == 5?'gg':'xx'}",
  "for($2.data,$val.id in const['aa','bb']? $$.status = $val.data.status ; $$::set($val.id,$val.data) : ( !$$.status? $$.status = $val.status ; $$.result = $val.data : _))",
  "hddef =  {name: 'hello', age:36, bios: {name:'atm', age:34},'fail': name or 1}",
  "$2::delete('data')",
  "assd = adf or names or 4",
  "assd2 = adf or name or 4",
  "str2 = \"helloworld\" ",
  "$map_to_str = {_ss}=>( _sb=str_builder(); for(_ss, _sb::write($key,'=',$val,';')); _sb::string()::trim_right(';') )",
  "mapstr1 = $map_to_str(kv.c)",
  "mapstr1 = $map_to_str(kv.c)",
  "callbool = nil::boolean()",
  "$.arr = slice_new(5)",
  "$.arr[3] = 'bb'",
  "$->arr[1] = 'bb'",
  "$->sbss = str_builder()::write('he')::write('ll')->write('o')::string()",
  "arr_set[0][0]='1';arr_set[0][1]='1';arr_set[1][0]='1';arr_set[1][1]='1'",
  "arr_set2 = [['1','1'],['1','1']]",
  "map_set1['name']='ns';map_set1['age'] = '3'",
  "map_set2['class']='ns';map_set2['pl'] = '3'",
  "map_set2['name']['name']='ns';map_set2['name']['age']='3'",
  "$index=2;$def[$index]='2'",
  "strfmt = '${name}_${name}'#",
  "kv::get('c')['e']='x2'",
  "$add = {$a,$b}=> $a + $b",
  "lmadd = $add(3,4)",
  "sbb=str_builder(); {d:'x'}::for({k,v}=>sbb::write(k,v));sbbs=sbb::string()",
  "return(1)",
  "cbg=2"
]
`
	o, err := ParseFromJSONStr(scpt)
	if err != nil {
		panic(err)
	}
	c := NewContext(map[string]any{
		"kv": map[string]any{
			"kv.a": "a",
			"kv.b": "b",
			"c": map[string]any{
				"d": "x",
			},
		},
		"$2": map[string]any{
			"data": []any{
				map[string]any{
					"id":     "aa",
					"data":   "data1",
					"status": 1,
				},
				map[string]any{
					"id":     "",
					"data":   "data2",
					"status": 2,
				},
			},
		},
	})
	c.table["$$"] = map[string]any{}
	err = c.Exec(o)
	if err != nil {
		panic(err)
	}
	assertEqual(t, c, "name", "hello")
	assertEqual(t, c, ("age"), float64(5))
	assertEqual(t, c, ("gender"), "a1")
	assertEqual(t, c, ("gender2"), "a2")
	assertEqual(t, c, ("ageset"), float64(1))
	assertEqual(t, c, ("ageset2"), float64(2))
	assertEqual(t, c, ("$.kv\\.a"), "a")
	assertEqual(t, c, ("$.kv\\.b"), "b")
	assertEqual(t, c, ("header.name"), "nn")
	assertEqual(t, c, ("cbool"), true)
	assertEqual(t, c, ("cb2"), "hello")
	assertEqual(t, c, ("cb3"), "yes")
	assertEqual(t, c, ("sname"), "1")
	assertEqual(t, c, ("ssname"), "12")
	assertEqual(t, c, ("fmt"), "hello_5")
	assertEqual(t, c, ("fmt2"), "${name}_5")
	assertEqual(t, c, ("fmt3"), "he\nhe")
	assertEqual(t, c, ("fmt5"), "is_hello")
	assertEqual(t, c, ("fmt6"), "age_2")
	assertEqual(t, c, ("$.calc_func"), "cbc.business.total")
	assertEqual(t, c, ("$.cset"), "1")
	assertEqual(t, c, ("e"), float64(4))
	assertEqual(t, c, ("d"), float64(4))
	assertEqual(t, c, ("f"), float64(4))
	assertEqual(t, c, ("g"), float64(4))
	assertEqual(t, c, ("$.cset1"), "1")
	assertEqual(t, c, ("cbg"), nil)
	assertEqual(t, c, ("$.top"), float64(6))
	assertEqual(t, c, ("$.top2"), float64(4))
	assertEqual(t, c, ("$.route"), "default")
	assertEqual(t, c, ("$.route2"), "/abc")
	assertEqual(t, c, ("$.route3"), "/r3")
	assertEqual(t, c, ("$.route4"), "/r5")
	assertEqual(t, c, ("$.route5"), "/r3")
	assertEqual(t, c, ("haxp"), true)
	assertEqual(t, c, ("haxpf"), false)
	assertEqual(t, c, ("cfa"), "a")
	assertEqual(t, c, ("cf2"), "x")
	assertEqual(t, c, ("sb"), "helloworld")
	assertEqual(t, c, ("haxpn"), true)
	assertEqual(t, c, ("haxpnn"), false)
	assertEqual(t, c, ("b64"), "hello")
	assertEqual(t, c, ("nilt"), "nil")
	assertEqual(t, c, ("m1"), float64(1))
	assertEqual(t, c, ("m2"), float64(2))
	assertEqual(t, c, ("m3"), float64(3))
	assertEqual(t, c, ("assd"), float64(4))
	assertEqual(t, c, ("assd2"), "hello")
	assertEqual(t, c, ("str2"), "helloworld")
	assertEqual(t, c, ("mapstr1"), "d=x")
	assertEqual(t, c, ("callbool"), false)
	assertEqual(t, c, ("$.arr[3]"), "bb")
	assertEqual(t, c, ("$.arr[1]"), "bb")
	assertEqual(t, c, ("$.sbss"), "hello")
	assertEqual(t, c, ("$def[2]"), "2")
	assertEqual(t, c, ("map_set1.name"), "ns")
	assertEqual(t, c, ("map_set1.age"), "3")
	assertEqual(t, c, ("map_set2.name.name"), "ns")
	assertEqual(t, c, ("map_set2.name.age"), "3")
	assertEqual(t, c, ("map_set2.class"), "ns")
	assertEqual(t, c, ("map_set2.pl"), "3")
	assertEqual(t, c, ("strfmt"), "hello_hello")
	assertEqual(t, c, ("kv.c.e"), "x2")
	assertEqual(t, c, ("lmadd"), float64(7))
	assertEqual(t, c, ("sbbs"), "dx")
	assertDeepEqual(t, c, ("$$"), c.GetByJp("$2"))
	assertDeepEqual(t, c, ("arr_set"), c.GetByJp("arr_set2"))
	assertDeepEqual(t, c, ("arr_set"), []any{[]any{"1", "1"}, []any{"1", "1"}})
	//fmt.Println(c.table)
	//fmt.Println(c.GetReturn())
	//bs, _ := json.MarshalIndent(c.table, "", "  ")
	//fmt.Println(string(bs))

	c.GetReturn()
}

func assertEqual(t *testing.T, c *Context, k string, b any) {
	a := c.GetByJp(k)
	if a != b {
		t.Errorf("FAILED: %s %v != %v", k, a, b)
	}
}
func assertEqual2(t *testing.T, a any, b any) {
	if a != b {
		t.Errorf("FAILED: %v != %v", a, b)
	}
}

func assertDeepEqual(t *testing.T, c *Context, k string, b any) {
	a := c.GetByJp(k)
	if !reflect.DeepEqual(a, b) {
		t.Errorf("FAILED: %s %v != %v", k, a, b)
	}
}

func TestParser(t *testing.T) {
	psr := &strparser{
		str: []rune("hello world$${$.name}$${age()}\\${}"),
	}
	err := psr.parser()
	if err != nil {
		panic(err)
	}

}

func BenchmarkStrVal(b *testing.B) {
	b.ReportAllocs()
	convertToError(0)

	s := &stringFmtVal{
		vals: []Val{
			&constraint{value: "strring"},
			&constraint{value: "strring2"},
			&constraint{value: "strring3"},
		},
	}
	for i := 0; i < b.N; i++ {
		s.Val(nil)
	}
}

var (
	started = 0
)

func BenchmarkStal(b *testing.B) {

	//e, err := ParseFromJSONStr(` "name2 = 'sms_${add(2,3)}.1'"`)
	e, err := ParseFromJSONStr(`

[
{
	"for":"k,v in $.data",
	"do":[
		{
			"if":"in(v.data_id,'res','okl')",
			"then":[
				"set($,'${v.data_id}_status',v.status)",
				"set($,v.data_id,v.data)"
			],
			"else":[
				"if(!$.result,($.result=v.data ) && ($.status=v.status))"
			 ]
		}
	]
}
]

`)
	//e, err := ParseFromJSONStr(` "$.calc_func = $.channel == 'vms' ? '${$.channel}.tokens.total' : ( $.channel == 'cbc' ? '${$.channel}.business.total' : 'business.total' ) "`)

	if err != nil {
		panic(err)
	}
	b.ReportAllocs()
	c := NewContext(map[string]any{
		"$": map[string]any{
			"data": []any{
				map[string]any{
					"data_id": "res",
					"data":    "result_new",
					"status":  "1",
				},
				map[string]any{
					"data_id": "",
					"data":    "result_old",
					"status":  "2",
				},
			},
		},
	})
	for i := 0; i < b.N; i++ {
		c.Exec(e)
	}
	fmt.Println(c.table)

	bs, _ := json.MarshalIndent(c.table, "", "  ")
	fmt.Println(string(bs))
}

func BenchmarkSP(b *testing.B) {
	b.ReportAllocs()
	tab := map[string]any{
		"name": "hello",
		"age":  "5",
	}
	for i := 0; i < b.N; i++ {
		tab["name2"] = fmt.Sprintf("%s:%s", tab["name"], tab["age"])
	}
}

func BenchmarkStringOf(b *testing.B) {
	var a string
	for i := 0; i < b.N; i++ {
		a = StringOf("hello world")
	}
	_ = a
}

func BenchmarkMap(b *testing.B) {
	mm := map[reflect.Type]any{
		reflect.TypeOf(map[string]any{}): map[string]any{},
	}
	for i := 0; i < b.N; i++ {
		_ = mm[reflect.TypeOf(mm)]
	}
}

func BenchmarkMap2(b *testing.B) {
	p := new(string)
	mm := map[*string]any{
		p: 1,
	}
	for i := 0; i < b.N; i++ {
		_ = mm[p]
	}
}

func TestDoc(t *testing.T) {

	//docsObj := []string{}
	//for _, m := range objFuncMap {
	//	for _, o := range m {
	//
	//		docsObj = append(docsObj, fmt.Sprintf("%s::%s\n", o.typeI, o.doc))
	//	}
	//}
	//
	//sort.Strings(docsObj)
	//glb := []string{}
	//for _, i := range funtables {
	//	glb = append(glb, fmt.Sprintf("%s()  args: %d\n", i.name, i.argsNum))
	//}
	//sort.Strings(glb)
	//
	//bs, err := ioutil.ReadFile("readme.tlp.md")
	//if err != nil {
	//	panic(err)
	//}
	//tp, err := template.New("").Parse(string(bs))
	//if err != nil {
	//	panic(err)
	//}
	//out := &bytes.Buffer{}
	//tp.Execute(out, map[string]interface{}{
	//	"global_func": strings.Join(glb, ""),
	//	"obj_func":    strings.Join(docsObj, ""),
	//})
	//
	//ioutil.WriteFile("readme.md", out.Bytes(), 0644)
}

func TestMath(t *testing.T) {
	e, err := ParseFromJSONStr(`
[
"a = 5",
"b = 6",
"g = 2.2",
"c = a ^ b",
"d = a / b",
"e = a % b",
"f = a+b*a - b",
"h = (a^b+b)*g/(a-b)",
"k = (-12)+a*b/2+1.2"
]
`)
	if err != nil {
		panic(err)
	}

	c := NewContext(map[string]any{})
	err = c.Exec(e)
	if err != nil {
		panic(err)
	}
	a := 5.0
	b := 6.0
	g := 2.2
	assertEqual(t, c, "a", 5.0)
	assertEqual(t, c, "b", 6.0)
	assertEqual(t, c, "c", math.Pow(5, 6))
	assertEqual(t, c, "d", 5.0/6.0)
	assertEqual(t, c, "e", float64(5%6))
	assertEqual(t, c, "h", (math.Pow(a, b)+b)*g/(a-b))
	assertEqual(t, c, "k", (-12)+a*b/2+1.2)
}

func TestAccess(t *testing.T) {
	e, err := ParseFromJSONStr(`
[
"a.b.c = 1",
"b[1] = 1",
"b[2] = 1.1",
"b[0] = 1.2",
"b['x'] = 22",
"a.b['1']=3"
]
`)
	if err != nil {
		panic(err)
	}

	c := NewContext(map[string]any{})
	err = c.Exec(e)
	if err != nil {
		panic(err)
	}

	assertDeepEqual(t, c, ("a"), map[string]any{
		"b": map[string]any{
			"c": 1.0,
			"1": 3.0,
		},
	})
	assertDeepEqual(t, c, "b", []any{1.2, 1.0, 1.1})

}

/*
 */

type CustomData struct {
	Name string
	Age  int
}

func (c *CustomData) SetField(ctx *Context, name string, val any) {
	switch name {
	case "name":
		c.Name = val.(string)
	case "age":
		c.Age = int(NumberOf(val))
	}
}

func (c *CustomData) GetField(ctx *Context, key string) any {
	switch key {
	case "name":
		return c.Name
	case "age":
		return float64(c.Age)
	}
	return nil
}

func TestAccessStruct(t *testing.T) {

	data := &CustomData{
		Name: "111",
		Age:  22,
	}
	e, err := ParseFromJSONStr(`
[
"name = data.name",
"age = data.age",
"data.name = '222'",
"data.age = 33"
]
`)
	if err != nil {
		panic(err)
	}

	c := NewContext(map[string]any{
		"data": data,
	})
	err = c.Exec(e)
	if err != nil {
		panic(err)
	}

	assertEqual(t, c, "name", "111")
	assertEqual(t, c, "age", 22.0)
	assertDeepEqual(t, c, "data", &CustomData{
		Name: "222",
		Age:  33,
	})
}

func TestAllFunc(t *testing.T) {
	data := &CustomData{
		Name: "111",
		Age:  22,
	}
	RegisterFunc("reterr", func(ctx *Context, args ...Val) any {
		return &Result{
			Err:  "return error",
			Data: "hello world",
		}
	}, 0)
	e, err := ParseFromJSONStr(`
[
"mapp = const {name:'he'}",
"abc = ass.catch()",
"e  =  err.catch()",
"dt = data.type()",
"rerr = reterr().catch()",
"rest = result.catch()",
"resterr = result.unwrap()"

]
`)
	if err != nil {
		panic(err)
	}

	c := NewContext(map[string]any{
		"data": data,
		"ass":  "name",
		"err": &Error{
			Err: fmt.Errorf("this is error"),
		},
		"result": &Result{
			Data: "hello world",
			Err:  "result err",
		},
	})
	err = c.Exec(e)
	if err != nil {
		re, ok := err.(*RuntimeError)
		if ok {
			panic("runtime:" + re.Error())
		}
	}
	assertEqual(t, c, "abc", "name")
	assertEqual(t, c, "e", nil)
	assertEqual(t, c, "dt", reflect.TypeOf(data).String())
	assertEqual(t, c, "rest", "hello world")
	assertDeepEqual(t, c, "resterr", newError("result err"))
	assertDeepEqual(t, c, "mapp", map[string]any{"name": "he"})
	//assertEqual(t, c, "name", "111")
	//assertEqual(t, c, "age", 22.0)
	//assertDeepEqual(t, c, "data", &CustomData{
	//	Name: "222",
	//	Age:  33,
	//})
}

func TestGG(t *testing.T) {
	for i := 0; i < 4e9; i++ {

	}
}

func TestHH(t *testing.T) {
	fmt.Println(calcHash("has_suffix") == calcHash("has_prefix"))
}
func deep(n int) {
	if n == 0 {
		panic("x")
	}
	deep(n - 1)
}
func exdc(f func()) (res any) {
	defer func() {
		res = recover()
	}()
	f()
	return nil
}

func BenchmarkPanic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		exdc(func() {

		})
	}
}
