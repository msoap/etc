package main

import (
	"fmt"
	"log"

	"github.com/valyala/fasthttp"
)

// ----------------------------------------------------------------------------
func main() {
	fmt.Println("Start fasthttp on 8080")
	responseBytes := []byte("Hello world from Go/Fasthttp/9012345678901234567890123456789012345678901234567890123456789/12")

	err := fasthttp.ListenAndServe(":8080", func(ctx *fasthttp.RequestCtx) {
		_, err := ctx.Write(responseBytes)
		if err != nil {
			log.Print(err)
			return
		}
	})
	if err != nil {
		log.Fatal(err)
	}
}
