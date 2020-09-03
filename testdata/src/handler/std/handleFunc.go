package std

import (
	"fmt"
	"net/http"
)

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

func f2() {
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
