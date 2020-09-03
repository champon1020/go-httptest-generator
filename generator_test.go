package generator_test

import (
	"testing"

	generator "github.com/champon1020/go-httptest-generator"
)

func TestGenerateTest(t *testing.T) {
	handlerInfo := &generator.HandlerInfo{
		Pkg:    "a",
		File:   "a",
		Name:   "handler",
		URL:    "/test",
		Method: "GET",
	}
	generator.GenerateTest(handlerInfo)
}
