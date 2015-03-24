// read all integers from stdin and out summation of it
// echo "123 456 789" | sum-big-integers
package main

import (
	"bufio"
	"fmt"
	"log"
	"math/big"
	"os"
	"regexp"
)

func main() {
	re := regexp.MustCompile(`\d+`)
	i := new(big.Int)
	sum := new(big.Int)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		for _, integer_str := range re.FindAllString(scanner.Text(), -1) {
			i.SetString(integer_str, 10)
			sum.Add(sum, i)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Println("reading standard input error:", err)
	}

	fmt.Println(sum)
}
