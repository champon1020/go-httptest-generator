package std

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

type anyHandler struct{}

func (a *anyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	fmt.Fprintf(w, "hello world")
}

var A = &AnyHandler{}
var a = &AnyHandler{}
var aa = &anyHandler{}

func f1() {
	var AA = &AnyHandler{}

	http.Handle("/handle1", new(AnyHandler)) // want "Handle /handle1 POST AnyHandler"
	http.Handle("/handle2", new(anyHandler)) // Ignore
	http.Handle("/handle3", A)               // want "Handle /handle3 POST AnyHandler"
	http.Handle("/handle4", AA)              // Ignore
	http.Handle("/handle5", a)               // want "Handle /handle5 POST AnyHandler"
	http.Handle("/handle6", aa)              // Ignore
}
