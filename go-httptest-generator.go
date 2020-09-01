package generator

import (
	"golang.org/x/tools/go/analysis"
)

const doc = "go-httptest-generator is ..."

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "go-httptest-generator",
	Doc:  doc,
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// some process
	return nil, nil
}
