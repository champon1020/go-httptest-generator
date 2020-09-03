package a

import (
	"fmt"
	"net/http"
)

type AnyHandler struct{}

func (a *AnyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	fmt.Fprintf(w, "hello world")
}

type AnyHandler2 struct{}

func (a *AnyHandler2) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	fmt.Fprintf(w, "hello world")
}

func main() {
	http.Handle("/handle1", new(AnyHandler)) // want "Handle /handle1 POST"

	anyHandler := &AnyHandler{}
	http.Handle("/handle2", anyHandler) // want "Handle /handle2 GET"

	http.Handle("/handlerFunc1", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // want "Handle /handlerFunc1 POST"
		if r.Method != "POST" {
			return
		}
		fmt.Fprintf(w, "hello world")
	}))

	http.HandleFunc("/handleFunc1", func(w http.ResponseWriter, r *http.Request) { // want "HandleFunc /handleFunc1 POST"
		if r.Method != "POST" {
			return
		}
		fmt.Fprintf(w, "hello world")
	})
}
