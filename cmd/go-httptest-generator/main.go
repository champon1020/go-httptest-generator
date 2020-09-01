package main

import (
	generator "github.com/champon1020/go-httptest-generator"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(generator.HandlerAnalyzer) }
