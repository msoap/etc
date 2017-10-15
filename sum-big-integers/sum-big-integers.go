/*
read all integers from stdin and out summation of it
echo "123 456 789" | sum-big-integers

Install:
	go get -u github.com/msoap/etc/sum-big-integers

*/
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"regexp"
	"runtime"
)

const bufferLen = 65536

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
func numbersSummator(input <-chan []string, output chan<- *big.Int) {
	i := new(big.Int)
	sum := new(big.Int)

	for line := range input {
		for _, integerStr := range line {
			i.SetString(integerStr, 10)
			sum.Add(sum, i)
		}
	}

	output <- sum
}

//-------------------------------------------------------------------
func main() {
	reader := bufio.NewReader(os.Stdin)
	buffer := make([]byte, bufferLen)
	prevStrNum := ""

	cntParallels := runtime.NumCPU()

	input := make(chan []string, cntParallels)
	output := make(chan *big.Int)
	re := regexp.MustCompile(`\d+`)

	for i := 1; i <= cntParallels; i++ {
		go numbersSummator(input, output)
	}

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalln("reading standard input error:", err)
		} else if n > 0 {
			buffer = buffer[:n]
			numsInLine := re.FindAllString(string(buffer), -1)

			// add previous num
			if prevStrNum != "" && isNumericSymbol(buffer[0]) {
				numsInLine[0] = prevStrNum + numsInLine[0]
			} else if prevStrNum != "" {
				input <- []string{prevStrNum}
			}

			// save last num
			if isNumericSymbol(buffer[n-1]) {
				prevStrNum = numsInLine[len(numsInLine)-1]
				numsInLine = numsInLine[:len(numsInLine)-1]
			} else {
				prevStrNum = ""
			}

			if len(numsInLine) > 0 {
				input <- numsInLine
			}
		}
	}

	if prevStrNum != "" {
		input <- []string{prevStrNum}
	}

	close(input)

	result := new(big.Int)
	for i := 1; i <= cntParallels; i++ {
		result.Add(result, <-output)
	}

	fmt.Println(result)
}
