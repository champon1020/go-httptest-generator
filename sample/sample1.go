package sample

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
var A2 = &anyHandler{}
var a = &AnyHandler{}
var a2 = &anyHandler{}

func f1() {
	var A3 = &AnyHandler{}

	http.Handle("/handle1", new(AnyHandler)) // want "Handle /handle1 POST AnyHandler"
	http.Handle("/handle2", new(anyHandler)) // Ignore
	http.Handle("/handle3", A)               // want "Handle /handle3 POST A"
	http.Handle("/handle4", A2)              // want "Handle /handle4 POST A2"
	http.Handle("/handle5", A3)              // Ignore
	http.Handle("/handle6", a)               // want "Handle /handle6 POST AnyHandler"
	http.Handle("/handle7", a2)              // Ignore
}
