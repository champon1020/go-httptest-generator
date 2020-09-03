package a

import (
	"fmt"
	"net/http"
)

func test() {
	fmt.Println("hoge")
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
}

func main() {
	http.HandleFunc("/test1", func(w http.ResponseWriter, r *http.Request) { // want "HandleFunc /test1 POST"
		if r.Method != "POST" {
			return
		}
		fmt.Fprintf(w, "hello world")
	})

	http.HandleFunc("/test2", func(w http.ResponseWriter, _ *http.Request) { // want "HandleFunc /test2 GET"
		fmt.Fprintf(w, "hello world")
	})

	http.HandleFunc("/test3", index) // want "HandleFunc /test3 GET"

	test()
}
