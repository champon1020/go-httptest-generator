package std

import (
	"fmt"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
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

var IndexVar = Index
var indexVar = Index
var indexVar_ = index

func f2() {
	var IndexVar_ = Index

	http.HandleFunc("/handleFunc1", Index)     // want "HandleFunc /handleFunc1 POST Index"
	http.HandleFunc("/handleFunc2", index)     // Ignore
	http.HandleFunc("/handleFunc3", IndexVar)  // want "HandleFunc /handleFunc3 POST Index"
	http.HandleFunc("/handleFunc4", IndexVar_) // Ignore
	http.HandleFunc("/handleFunc5", indexVar)  // want "HandleFunc /handleFunc5 POST Index"
	http.HandleFunc("/handleFunc6", indexVar_) // Ignore
}
