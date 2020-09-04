package std

import (
	"fmt"
	"net/http"
)

var H = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	fmt.Fprintf(w, "hello world")
})

var h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	fmt.Fprintf(w, "hello world")
})

func f3() {
	var h2 = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}
		fmt.Fprintf(w, "hello world")
	})

	http.Handle("/handlerFunc1", H)                           // want "Handle /handlerFunc1 POST H"
	http.Handle("/handlerFunc2", h)                           // Ignore
	http.Handle("/handlerFunc3", h2)                          // Ignore
	http.Handle("/handlerFunc4", http.HandlerFunc(Index))     // want "Handle /handlerFunc4 POST Index"
	http.Handle("/handlerFunc5", http.HandlerFunc(index))     // Ignore
	http.Handle("/handlerFunc6", http.HandlerFunc(IndexVar))  // want "Handle /handlerFunc6 POST Index"
	http.Handle("/handlerFunc7", http.HandlerFunc(indexVar))  // want "Handle /handlerFunc7 POST Index"
	http.Handle("/handlerFunc8", http.HandlerFunc(indexVar2)) // Ignore

	http.Handle("/handlerFunc9", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // Ignore
		if r.Method != "POST" {
			return
		}
		fmt.Fprintf(w, "hello world")
	}))
}
