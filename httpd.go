package main

import (
    "fmt"
    "log"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello world from Go/012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789")
}

func main() {
    http.HandleFunc("/", handler)
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal(err)
    }
}
