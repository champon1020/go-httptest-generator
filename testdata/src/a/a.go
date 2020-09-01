package a

import (
	"fmt"
	"net/http"
)

func test() {
	fmt.Println("hoge")
}

func main() {
	http.HandleFunc("/test1", func(w http.ResponseWriter, r *http.Request) { // want "http.HandleFunc with /test1"
		fmt.Fprintf(w, "hello world")
	})

	http.HandleFunc("/test2", func(w http.ResponseWriter, r *http.Request) { // want "http.HandleFunc with /test2"
		fmt.Fprintf(w, "hello world")
	})

	http.HandleFunc("/test3", func(w http.ResponseWriter, r *http.Request) { // want "http.HandleFunc with /test3"
		fmt.Fprintf(w, "hello world")
	})

	test()
}
