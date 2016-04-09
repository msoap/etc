package main

import (
	"log"
	"net/http"
)

func main() {
	responseBytes := []byte("Hello world from Go/012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(responseBytes)
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
