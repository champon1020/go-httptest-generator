package std

import (
	"fmt"
	"net/http"
)

// AnyHandler is the handler that is exported.
type AnyHandler struct{}

func (a *AnyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	fmt.Fprintf(w, "hello world")
}

// anyHandler is the handler that is not exported.
type anyHandler struct{}

func (a *anyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	fmt.Fprintf(w, "hello world")
}

// Handler instances that rhs is exported.
var (
	H = new(AnyHandler)
	h = new(AnyHandler)
)

// Handler instances that rhs is not exported.
var (
	H2 = new(anyHandler)
	h2 = new(anyHandler)
)

// Handler instances that is created by http.HandlerFunc and handler function is exported.
var (
	H3 = http.HandlerFunc(HFunc)
	h3 = http.HandlerFunc(HFunc)
)

// Handler instances that is created by http.HandlerFunc and handler function is not exported.
var (
	H4 = http.HandlerFunc(hFunc)
	h4 = http.HandlerFunc(hFunc)
)

func f1() {
	var (
		H5 = new(AnyHandler)
		h5 = new(AnyHandler)
	)

	http.Handle("/handle", new(AnyHandler))            // want "Handle /handle POST AnyHandler"
	http.Handle("/handle", new(anyHandler))            // Ignore
	http.Handle("/handle", H)                          // want "Handle /handle POST H2"
	http.Handle("/handle", h)                          // Ignore
	http.Handle("/handle", H2)                         // want "Handle /handle POST H3"
	http.Handle("/handle", h2)                         // Ignore
	http.Handle("/handle", H3)                         // want "Handle /handle POST H4"
	http.Handle("/handle", h3)                         // Ignore
	http.Handle("/handle", H4)                         // want "Handle /handle POST H5"
	http.Handle("/handle", h4)                         // Ignore
	http.Handle("/handle", http.HandlerFunc(HFunc))    // want "Handle /handle POST HFunc"
	http.Handle("/handle", http.HandlerFunc(hFunc))    // Ignore
	http.Handle("/handle", http.HandlerFunc(HFuncLit)) // want "Handle /handle POST HFuncLit"
	http.Handle("/handle", http.HandlerFunc(hFuncLit)) // Ignore
	http.Handle("/handle", H5)                         // Ignore
	http.Handle("/handle", h5)                         // Ignore

	http.Handle("/handle", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})) // Ignore
}
