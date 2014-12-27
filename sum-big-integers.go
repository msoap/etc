// read all integers from stdin and out summation of it
// echo "123 456 789" | sum-big-integers
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"regexp"
)

func main() {
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalln("read from stdin failed")
	}

	re := regexp.MustCompile(`\d+`)
	i := new(big.Int)
	sum := new(big.Int)

	for _, integer_str := range re.FindAllString(string(stdin), -1) {
		fmt.Sscan(integer_str, i)
		sum.Add(sum, i)
	}
	fmt.Println(sum)
}
