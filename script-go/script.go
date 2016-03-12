//usr/bin/env go run $0 "$@"; exit
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello world!")
	cwd, _ := os.Getwd()
	fmt.Println("cwd:", cwd)
	fmt.Println("args:", os.Args[1:])
}
