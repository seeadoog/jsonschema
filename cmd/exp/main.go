package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	expr2 "github.com/seeadoog/jsonschema/v2/expr"
	"io"
	"io/ioutil"
	"os"
	"path"
)

func readData() any {
	data, _ := io.ReadAll(os.Stdin)
	var i any
	err := json.Unmarshal(data, &i)
	if err != nil {
		panic(err)
	}
	return i
}

func initFunc() {
	expr2.RegisterFunc("read", expr2.FuncDefine1(func(a string) any {
		bs, err := ioutil.ReadFile(a)
		if err != nil {
			return &expr2.Error{Err: err.Error()}
		}
		var i any
		err = json.Unmarshal(bs, &i)
		if err != nil {
			return &expr2.Error{Err: err.Error()}
		}
		return i

	}), 1)

	expr2.RegisterFunc("import", func(ctx *expr2.Context, arg ...expr2.Val) any {
		v, err := importVal(expr2.StringOf(arg[0].Val(ctx)))
		if err != nil {
			return err
		}
		return v.Val(ctx)
	}, 1)
}

func importVal(f string) (expr2.Val, error) {
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, &expr2.Error{Err: fmt.Sprintf("failed to import file %s %v", f, err)}
	}
	val, err := expr2.ParseValue(string(bs))
	if err != nil {
		return nil, &expr2.Error{
			Err: fmt.Sprintf("failed to parse value %s %v", string(bs), err),
		}
	}
	return val, nil
}

func importFromENV(ctx *expr2.Context) {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	dd := path.Join(home, "/.explib")
	dir, err := os.ReadDir(dd)
	if err != nil {
		return
	}
	for _, fi := range dir {
		if fi.IsDir() {
			continue
		}
		name := fi.Name()
		v, err := importVal(path.Join(dd, name))
		if err != nil {
			panic(fmt.Sprintf("failed to import file %s %v", name, err))
		}
		v.Val(ctx)
	}
}

func main() {
	initFunc()
	file := ""
	expr := ""
	start := ""
	end := ""
	flag.StringVar(&file, "f", "", "file to parse")
	flag.StringVar(&expr, "e", "", "expression to parse")
	flag.StringVar(&start, "st", "", "start expression")
	flag.StringVar(&end, "ed", "", "end expression")

	flag.Parse()
	c := expr2.NewContext(map[string]any{})

	importFromENV(c)
	if file != "" {
		f, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		sc := bufio.NewScanner(f)
		e, err := expr2.ParseValue(expr)
		if err != nil {
			panic(err)
		}
		var o any

		if start != "" {
			st, err := expr2.ParseValue(start)
			if err != nil {
				panic("invalid start expression:" + err.Error())
			}
			o = st.Val(c)
		}

		sc.Buffer(make([]byte, 1024*1024*4), 1024*1024*4)
		idx := 0
		for sc.Scan() {
			txt := sc.Text()
			if len(txt) == 0 {
				continue
			}
			var i any
			err := json.Unmarshal(expr2.ToBytes(txt), &i)
			if err != nil {
				i = txt
			}
			c.Set("$", i)
			c.Set("$idx", float64(idx))
			idx++
			o = e.Val(c)
		}

		if end != "" {
			ed, err := expr2.ParseValue(end)
			if err != nil {
				panic("invalid end expression:" + err.Error())
			}
			o = ed.Val(c)
		}
		if o != nil {
			bs, _ := json.MarshalIndent(o, "", "    ")
			fmt.Println(string(bs))
		}

	} else {

		c.Set("$", readData())

		e, err := expr2.ParseValue(os.Args[1])
		if err != nil {
			panic(err)
		}

		o := e.Val(c)
		if o != nil {
			bs, _ := json.MarshalIndent(o, "", "    ")
			fmt.Println(string(bs))
		}

	}

}
