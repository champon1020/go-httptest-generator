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

func Index2(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	fmt.Fprintf(w, "hello world")
}

func index3(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	fmt.Fprintf(w, "hello world")
}

func f2() {
	http.HandleFunc("/handleFunc0", func(w http.ResponseWriter, r *http.Request) { // Ignore
		if r.Method != "POST" {
			return
		}
		fmt.Fprintf(w, "hello world")
	})

	http.HandleFunc("/handleFunc1", Index) // want "HandleFunc /handleFunc1 POST"

	{
		var Index = Index2
		http.HandleFunc("/handleFunc2", Index) // want "HandleFunc /handleFunc2 POST"
	}

	http.HandleFunc("/handleFunc0", index3) // Ignore
}
