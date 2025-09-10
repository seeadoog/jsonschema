package expr

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestPar(t *testing.T) {
	res, err := parseTokenizer("f('")
	//res, err := parseTokenizer("fa(a[0].ad,v,fb(b,fdd(1,2,'3'),c),'lla(sd)',bb,fc())")
	if err != nil {
		panic(err)
	}
	for i, re := range res {
		fmt.Println(i, string(byte(re.kind)), re.tkn)
	}

	v, er := parseTokenAsVal(res)
	if er != nil {
		panic(er)
	}
	fmt.Println(v)
}

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSetCond(tt.args.e); got != tt.want {
				t.Errorf("isSetCond() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpr(t *testing.T) {
	e, err := parseExpr("")
	if err != nil {
		panic(err)
	}

	ctx := &Context{
		table: map[string]any{
			"bs": []any{"1", "2"},
		},
	}
	err = e.Exec(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(ctx.table)
}
func TestVal(t *testing.T) {
	e, err := parseValueV("!eq(not(not(a+b+(c()))),b,c,'',eq(g,c,'a',string(1,2)))")
	if err != nil {
		panic(err)
	}

	ctx := &Context{
		table: map[string]any{
			"bs": []any{"1", "2"},
		},
	}
	n := e.Val(ctx)
	fmt.Println(n)
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
	RegisterDynamicFunc("test")
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

func TestForReg(t *testing.T) {
	fmt.Println(forRegexp.FindAllStringSubmatch("k , v in abc", -1))
}
func TestJSONScpt(t *testing.T) {
	RegisterDynamicFunc("add2")
	RegisterDynamicFunc("response.write")
	e, err := ParseFromJSONStr(`
[
"a.name=or(bss,time.format(time.now(),'2006-01-02 15:04:05'))",
{
	"if":"eqs(ass,'')",
	"then":["ac=or('',print(1,2,3,4))"]
},
{
	"for":"k,v in $.pss",
	"do":"print(k,v)"
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
"mm.name='5ds'",
"fff.ss=5",
"mm.age=number(3)",
"print('nameis:',mm.name)",
"ss[0]=1",
"ss[1]=2",
"ss=append(ss,3,4)",
"sss=slice.cut(ss,0,1)",
"sssa=str.join(ss,'|')",
"ss='asdnds'",
"ss3=ternary(mm.name,'true','false')",
"s5=hex.encode(md5(ss))",
"auths=sprintf('host: %v \ndate: %v','app.xxc.om',time.now())",
"sha=hex.encode(hmac.sha256(auths,'helloworld'))",
"header.X-Http-Name='ems'",
"hres=http.request('GET','http://172.30.209.27',nil,nil,1000)",
"gs=json.from(hres.body)",
"msg=gs.message",
"ges=d.d.d",
"bd=a=5",
{
	"switch":{
		"eq(mm.name,'5ds3')":"def.aa=1",
		"lt(mm.age,2)":"def.aa=2"
	},
	"default":["def.aa=5"]
}
]
`)
	if err != nil {
		panic(err)
	}
	ctx := &Context{
		table: map[string]any{
			"mm": map[string]any{},

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
		panic(err)
	}
	bs, _ := json.MarshalIndent(ctx.table, "", "  ")
	fmt.Println(string(bs))
	fmt.Println(ctx.Get("auths"))

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
func TestToSufix(t *testing.T) {

	fmt.Println(calcSuffix(toSufix("3+4+2*2/(2-1)*5+5")))
}

// abc/-
