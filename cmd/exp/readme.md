
## install

````shell
go install github.com/seeadoog/jsonschema/v2/exp@latest 
````

~/.explib 该目录中的文件会被预加载，exp 命令执行前会之前该目录的全部文件。
## Usage
```shell
exp -s 'print("hello")' # 处理个命令

cat jsonfile |exp '$.name'  #处理json 文件, 仅支持单个json

exp -f jsonfile_rows -e 'print($.name);a=a+1' -st 'a=1' -ed 'a' #处理json 文件，可以处理多行json ，st 和 ed 只会执行一次， e 则会每行都会执行。
```