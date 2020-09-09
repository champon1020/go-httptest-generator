package handler_test

import (
	"testing"

	"github.com/champon1020/go-httptest-generator/handler"
	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer is a test for handler.Analyzer.
func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, handler.Analyzer, "./src")
}
