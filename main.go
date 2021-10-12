package main

import (
	"fmt"
)

func main() {
	EvalString("(print \"hi from janet\")")
	fmt.Println("hi from go")
}