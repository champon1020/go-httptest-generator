package generator_test

import (
	"testing"

	generator "github.com/champon1020/go-httptest-generator"
)

func TestGenerateTest(t *testing.T) {
	handlerInfo := &generator.HandlerInfo{
		Pkg: generator.Package{
			Name: "a",
			Path: "path/to/a",
		},
		File:   "a.go",
		Name:   "index",
		URL:    "/test",
		Method: "GET",
	}
	generator.GenerateTest(handlerInfo)
}
