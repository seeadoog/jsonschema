package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	expr2 "github.com/seeadoog/jsonschema/v2/expr"
	"io"
	"os"
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

func main() {

	file := ""
	expr := ""
	start := ""
	end := ""
	flag.StringVar(&file, "f", "", "file to parse")
	flag.StringVar(&expr, "e", "", "expression to parse")
	flag.StringVar(&start, "st", "", "start expression")
	flag.StringVar(&end, "ed", "", "end expression")
	flag.Parse()

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
		c := expr2.NewContext(map[string]any{})
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

		e, err := expr2.ParseValue(os.Args[1])
		if err != nil {
			panic(err)
		}
		c := expr2.NewContext(map[string]any{
			"$": readData(),
		})

		o := e.Val(c)
		if o != nil {
			bs, _ := json.MarshalIndent(o, "", "    ")
			fmt.Println(string(bs))
		}

	}

}
