// read all integers from stdin and out summation of it
// echo "123 456 789" | sum-big-integers
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"regexp"
)

const buffer_len = 1000000

//-------------------------------------------------------------------
func isNumericSymbol(symbol byte) bool {
	switch symbol {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	}
	return false
}

//-------------------------------------------------------------------
// goroutine for summation of numbers in string
func numbersSummator(input chan []string, output chan *big.Int) {
	i := new(big.Int)
	sum := new(big.Int)

	for line := range input {
		for _, integer_str := range line {
			i.SetString(integer_str, 10)
			sum.Add(sum, i)
		}
	}

	output <- sum
}

//-------------------------------------------------------------------
func main() {
	reader := bufio.NewReader(os.Stdin)
	buffer := make([]byte, buffer_len)
	prev_str_num := ""

	input := make(chan []string)
	sum := make(chan *big.Int)
	re := regexp.MustCompile(`\d+`)
	go numbersSummator(input, sum)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalln("reading standard input error:", err)
		} else if n > 0 {
			buffer = buffer[:n]
			nums_in_line := re.FindAllString(string(buffer), -1)

			// add previous num
			if prev_str_num != "" && isNumericSymbol(buffer[0]) {
				nums_in_line[0] = prev_str_num + nums_in_line[0]
			} else if prev_str_num != "" {
				input <- []string{prev_str_num}
			}

			// save last num
			if isNumericSymbol(buffer[n-1]) {
				prev_str_num = nums_in_line[len(nums_in_line)-1]
				nums_in_line = nums_in_line[:len(nums_in_line)-1]
			} else {
				prev_str_num = ""
			}

			if len(nums_in_line) > 0 {
				input <- nums_in_line
			}
		}
	}

	if prev_str_num != "" {
		input <- []string{prev_str_num}
	}

	close(input)

	fmt.Println(<-sum)
}
