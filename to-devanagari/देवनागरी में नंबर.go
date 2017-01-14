/*
https://en.wikipedia.org/wiki/Devanagari#Numerals

usage:
	go run "देवनागरी में नंबर.go" "example: 2017, 01234567"

*/
package main

import (
	"fmt"
	"os"
	"strconv"
)

var toDevanagariMap = [...]rune{
	'०',
	'१',
	'२',
	'३',
	'४',
	'५',
	'६',
	'७',
	'८',
	'९',
}

func toDevanagari(in string) string {
	result := []rune{}
	for _, c := range in {
		d, err := strconv.Atoi(string(c))
		if err == nil && d >= 0 && d <= 9 {
			result = append(result, toDevanagariMap[d])
		} else {
			result = append(result, c)
		}
	}

	return string(result)
}

func main() {
	for _, arg := range os.Args[1:] {
		fmt.Println(toDevanagari(arg))
	}
}
