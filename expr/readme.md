


## Powerful Golang Expr Engine

```
a + b + c + (d*e)
name == '500'? 3 : 4 
aso = 4 
str_has_prefix(name,'123')
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
name.hex() # 调用string 的 hex() 方法，返回base16 编码

arr[3] = 5  
object.name = 5  。
object['name'] = 5 # 赋值
object::name = 5  #赋值
object->name = 5  #赋值  :: 和 -> 含义一致

object.name   #取值
object['name'] #取值
object::name #取值
object->name #取值

a=5 #注释，支持注释
'hello world'  # str 定义： 支持三种来包住str
`hello world`
"hello world"

'hello world ${name} time is ${time.now()}' # 变量嵌入字符串

dd = name == 'hello' ? 'abc' : dname or 'cname'  # 三元表达式
arr = [1,2,3,4,5] # 数组定义
obj = { name: '5' ,age: 8 } # map 定义

arr.for( {k,v} => print(k,v)) #lambda 表达式

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
*strings.Builder::write()*strings.Builder
[]interface {}::all(cond)[]any
[]interface {}::all(cond)[]any
[]interface {}::for(expr)
[]interface {}::get( float64)any
[]interface {}::json_str()string
[]interface {}::len()float64
[]interface {}::slice( float64, float64)any
[]interface {}::sort( any)any
[]uint8::base64()string
[]uint8::base64d()any
[]uint8::bytes()[]uint8
[]uint8::copy()[]uint8
[]uint8::hex()string
[]uint8::slice( float64, float64)[]uint8
float64::json_str()string
map[string]interface {}::delete( string)map[string]interface {}
map[string]interface {}::for(expr)
map[string]interface {}::get( string)any
map[string]interface {}::json_str()any
map[string]interface {}::len()float64
map[string]interface {}::set( string, any)map[string]interface {}
string::base64()string
string::base64d()any
string::bytes()[]uint8
string::contains( string)bool
string::fields()[]string
string::has( string)bool
string::has_prefix( string)bool
string::has_suffix( string)bool
string::hex()string
string::json_str()string
string::len()float64
string::md5()[]uint8
string::slice( float64, float64)string
string::trim( string)string
string::trim_left( string)string
string::trim_right( string)string
string::trim_space()string
time.Time::add_mill( float64)time.Time
time.Time::day()float64
time.Time::format( string)string
time.Time::hour()float64
time.Time::local()time.Time
time.Time::minute()float64
time.Time::month()float64
time.Time::second()float64
time.Time::sub( time.Time)float64
time.Time::unix()float64
time.Time::unix_micro()float64
time.Time::unix_mill()float64
time.Time::utc()time.Time
time.Time::year()float64
url.Values::encode()string
url.Values::get( string)string
url.Values::set( string, any)any


# 全局函数
add()  args: -1
add2()  args: 2
all()  args: 2
and()  args: -1
append()  args: -1
base64_decode()  args: 1
base64_encode()  args: 1
bool()  args: 1
boolean()  args: 1
bytes()  args: 1
catch()  args: 1
delete()  args: 2
div()  args: 2
eq()  args: 2
eqs()  args: 2
for()  args: 2
get()  args: 2
go()  args: 1
gt()  args: 2
gte()  args: 2
has_prefix()  args: 2
has_suffix()  args: 2
hex_decode()  args: 1
hex_encode()  args: 1
hmac_sha256()  args: 2
http_request()  args: 5
if()  args: -1
in()  args: -1
int()  args: 1
join()  args: -1
json_from()  args: 1
json_to()  args: 1
len()  args: 1
loop()  args: -1
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
recover()  args: 1
regexp_new()  args: 1
response_write()  args: 1
return()  args: -1
set()  args: 3
set_index()  args: 3
sha256()  args: 1
slice_cut()  args: 3
slice_init()  args: -1
slice_new()  args: -1
sprintf()  args: -1
str_builder()  args: 0
str_fields()  args: 1
str_join()  args: -1
str_split()  args: 3
str_to_lower()  args: 1
str_to_upper()  args: 1
str_trim()  args: 1
string()  args: 1
sub()  args: 2
ternary()  args: 3
time_format()  args: 2
time_from_unix()  args: 1
time_now()  args: 0
time_now_mill()  args: 0
time_parse()  args: 2
to_json_obj()  args: 1
to_json_str()  args: 1
type()  args: 1
unwrap()  args: 1
url_new_values()  args: 0


```