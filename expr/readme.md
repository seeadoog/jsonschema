


## Powerful Golang Expr Engine

```
a + b + c + (d*e)
name == '500'? 3 : 4 
aso = 4 
str.has_prefix(name,'123')
arr[0:3]
arr[:3]
arr[2:]
arr[3]
name = 5   # asign 5 to name 
a=4 ; b=5 ; c=6  #执行多个表达式，会返回最后一个表达式的值。
name or 'hello' 
a == 5 && b == 6
a == 5 || b == 6
a != 5 
!ok 
name::hex() # 调用string 的 hex() 方法，返回base16 编码

arr[3] = 5  
object.name = 5  # . 在变量中，会使用jsonpath 来取值，在函数名中则是普通字符，没有其他意义。
object['name'] = 5 # 赋值
object::name = 5  #赋值
object->name = 5  #赋值  :: 和 -> 含义一致

object.name   #取值
object['name'] #取值
object::name #取值
object->name #取值

a=5 #注释，支持注释
'hello world'
`hello world`
"hello world"

'hello world ${name} time is ${time.now()}' # 变量嵌入字符串

{ name :'hello',age :'5' ,smt: 3 + 3 ,'body':{}, class: pr_class or 'hello',frend:['join','jack'] }  # object define

[1,2,3,'4',{'name':3}]  # array define 
const arr = [1,3,4,5] #define const array value, 数组，map 中所有值必须是常量。同时，map，array 的值必须不能被改变，否则会出现运行时异常，并发读写等问题。 
all(arr,v => v > 5) #lambda
all(arr,{i,v} => i > 5) #lambda
```

### Usage
````go 
package main

import (
	"fmt"
	expr2 "github.com/seeadoog/jsonschema/v2/expr"
	"time"
)

func main() {
	expr2.RegisterDynamicFunc("get_cur_time", 0)
	expr, err := expr2.ParseValue(`'${name}_${get_cur_time()::format("2006-01-02 15:04:05")}'`)
	if err != nil {
		panic(err)
	}
	ctx := expr2.NewContext(map[string]any{
		"name": "hello",
	})
	ctx.SetFunc("get_cur_time", func(ctx *expr2.Context, args ...expr2.Val) any {
		return time.Now()
	})

	n := expr.Val(ctx)
	fmt.Println("result is:", n)

}

````


### 内置函数支持
```
#对象函数

*regexp.Regexp::match( string)bool
*strings.Builder::string()string
*strings.Builder::write()*strings.Builder
[]interface {}::get( float64)any
[]interface {}::len()float64
[]interface {}::slice( float64, float64)any
[]uint8::base64()string
[]uint8::base64d()[]uint8
[]uint8::bytes()[]uint8
[]uint8::copy()[]uint8
[]uint8::hex()string
[]uint8::slice( float64, float64)[]uint8
[]uint8::string()string
[]uint8::type()string
bool::string()string
bool::type()string
float64::string()string
float64::type()string
map[string]interface {}::delete( string)map[string]interface {}
map[string]interface {}::get( string)any
map[string]interface {}::len()float64
map[string]interface {}::set( string, any)map[string]interface {}
nil::bool()bool
nil::number()float64
nil::string()string
nil::type()string
string::base64()string
string::base64d()[]uint8
string::bytes()[]uint8
string::has_prefix( string)bool
string::has_suffix( string)bool
string::hex()string
string::len()float64
string::md5()[]uint8
string::slice( float64, float64)string
string::string()string
string::trim( string)string
string::trim_left( string)string
string::trim_right( string)string
string::trim_space()string
string::type()string
time.Time::add_mill( float64)time.Time
time.Time::day()float64
time.Time::format( string)string
time.Time::hour()float64
time.Time::local()time.Time
time.Time::month()float64
time.Time::sub( time.Time)float64
time.Time::unix()float64
time.Time::unix_nano()float64
time.Time::utc()time.Time
time.Time::year()float64
url.Values::encode()string
url.Values::get( string)string
url.Values::set( string, any)any


# 全局函数
add()  args: -1
all()  args: 2
and()  args: -1
append()  args: -1
base64.decode()  args: 1
base64.encode()  args: 1
bool()  args: 1
bytes()  args: 1
delete()  args: 2
div()  args: 2
eq()  args: 2
eqs()  args: 2
for()  args: 2
func()  args: 2
get()  args: 2
gt()  args: 2
gte()  args: 2
has_prefix()  args: 2
has_suffix()  args: 2
hex.decode()  args: 1
hex.encode()  args: 1
hmac.sha256()  args: 2
http.request()  args: 5
if()  args: -1
in()  args: -1
int()  args: 1
join()  args: -1
json.from()  args: 1
json.to()  args: 1
len()  args: 1
lt()  args: 2
lte()  args: 2
md5()  args: 1
mod()  args: 2
mul()  args: 2
neg()  args: 1
neq()  args: 2
neqs()  args: 2
new()  args: 0
not()  args: 1
number()  args: 1
or()  args: -1
orr()  args: 2
pow()  args: 2
print()  args: -1
regexp.new()  args: 1
return()  args: -1
set()  args: 3
set_index()  args: 3
sha256()  args: 1
slice.cut()  args: 3
slice.init()  args: -1
slice.new()  args: -1
sprintf()  args: -1
str.builder()  args: 0
str.join()  args: -1
str.split()  args: 3
str.to_lower()  args: 1
str.to_upper()  args: 1
str.trim()  args: 1
string()  args: 1
sub()  args: 2
ternary()  args: 3
time.format()  args: 1
time.from_unix()  args: 1
time.now()  args: 0
time.now_mill()  args: 0
type()  args: 1
url.new_values()  args: 0


```