package main

import (
	"github.com/champon1020/go-httptest-generator/handler"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	unitchecker.Main(handler.Analyzer)
}
