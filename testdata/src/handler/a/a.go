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
	if r.Method != "PUT" {
		return
	}
	fmt.Fprintf(w, "hello world")
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	fmt.Fprintf(w, "hello world")
}

func index2(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	fmt.Fprintf(w, "hello world")
}

func main() {
	http.Handle("/handle1", new(AnyHandler)) // want "Handle /handle1 POST"

	anyHandler := &AnyHandler{}
	http.Handle("/handle2", anyHandler) // want "Handle /handle2 POST"

	http.Handle("/handle3", new(AnyHandler2)) // want "Handle /handle3 PUT"

	{
		anyHandler := &AnyHandler2{}
		http.Handle("/handle4", anyHandler) // want "Handle /handle4 PUT"
	}

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

	http.HandleFunc("/handleFunc2", index) // want "HandleFunc /handleFunc2 POST"

	{
		var index = func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				return
			}
			fmt.Fprintf(w, "hello world")
		}
		http.HandleFunc("/handleFunc3", index) // want "HandleFunc /handleFunc3 POST"
	}
}
