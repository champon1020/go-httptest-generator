package std

import (
	"fmt"
	"net/http"
)

// HFunc is handler function that is exported.
func HFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	fmt.Fprintf(w, "hello world")
}

// hFunc is handler function that is not exported.
func hFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	fmt.Fprintf(w, "hello world")
}

// Function literal.
var (
	HFuncLit = func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}
		fmt.Fprintf(w, "hello world")
	}

	hFuncLit = func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}
		fmt.Fprintf(w, "hello world")
	}
)

// Handler function literal that rhs is exported.
var (
	HFuncLit2 = HFunc
	hFuncLit2 = HFunc
)

// Handler function literal that rhs is not exported.
var (
	HFuncLit3 = hFunc
	hFuncLit3 = hFunc
)

func f2() {
	var (
		HFuncLit4 = func(w http.ResponseWriter, r *http.Request) {}
		hFuncLit4 = func(w http.ResponseWriter, r *http.Request) {}
	)

	var (
		HFuncLit5 = HFunc
		hFuncLit5 = HFunc
	)

	var (
		HFuncLit6 = hFunc
		hFuncLit6 = hFunc
	)

	http.HandleFunc("/handleFunc", HFunc)     // want "HandleFunc /handleFunc POST HFunc"
	http.HandleFunc("/handleFunc", hFunc)     // Ignore
	http.HandleFunc("/handleFunc", HFuncLit)  // want "HandleFunc /handleFunc POST HFuncLit"
	http.HandleFunc("/handleFunc", hFuncLit)  // Ignore
	http.HandleFunc("/handleFunc", HFuncLit2) // want "HandleFunc /handleFunc POST HFuncLit2"
	http.HandleFunc("/handleFunc", hFuncLit2) // Ignore
	http.HandleFunc("/handleFunc", HFuncLit3) // want "HandleFunc /handleFunc POST HFuncLit3"
	http.HandleFunc("/handleFunc", hFuncLit3) // Ignore
	http.HandleFunc("/handleFunc", HFuncLit4) // Ignore
	http.HandleFunc("/handleFunc", hFuncLit4) // Ignore
	http.HandleFunc("/handleFunc", HFuncLit5) // Ignore
	http.HandleFunc("/handleFunc", hFuncLit5) // Ignore
	http.HandleFunc("/handleFunc", HFuncLit6) // Ignore
	http.HandleFunc("/handleFunc", hFuncLit6) // Ignore
}
