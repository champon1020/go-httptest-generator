package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("/test1", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // want "mux /test1 GET"
		fmt.Fprintf(w, "hello world")
	}))

	mux.Handle("/test2", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // want "mux /test2 POST"
		if r.Method != "POST" {
			return
		}
		fmt.Fprintf(w, "hello world")
	}))
}
