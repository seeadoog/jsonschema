package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024*10), 1024*1024*10)
	n := 0
	for sc.Scan() {
		sc.Text()

		n++
	}
	fmt.Println(n)

}
