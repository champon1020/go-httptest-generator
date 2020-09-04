package sample

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
var IndexVar2 = index
var indexVar = Index
var indexVar2 = index

func f2() {
	var IndexVar3 = Index
	var IndexVar4 = index

	http.HandleFunc("/handleFunc1", Index)     // want "HandleFunc /handleFunc1 POST Index"
	http.HandleFunc("/handleFunc2", index)     // Ignore
	http.HandleFunc("/handleFunc3", IndexVar)  // want "HandleFunc /handleFunc3 POST IndexVar"
	http.HandleFunc("/handleFunc4", IndexVar2) // want "HandleFunc /handleFunc4 POST IndexVar"
	http.HandleFunc("/handleFunc5", IndexVar3) // want "HandleFunc /handleFunc5 POST Index"
	http.HandleFunc("/handleFunc6", IndexVar4) // Ignore
	http.HandleFunc("/handleFunc7", indexVar)  // want "HandleFunc /handleFunc7 POST Index"
	http.HandleFunc("/handleFunc8", indexVar2) // Ignore
}
