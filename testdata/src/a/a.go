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
	http.HandleFunc("/test1", func(w http.ResponseWriter, r *http.Request) { // want "http.HandleFunc with /test1"
		if r.Method != "POST" {
			return
		}
		fmt.Fprintf(w, "hello world")
	})

	http.HandleFunc("/test2", index) // want "http.HandleFunc with /test2"

	test()
}
