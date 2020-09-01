package generator_test

import (
	"testing"

	generator "github.com/champon1020/go-httptest-generator"
	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer is a test for Analyzer.
func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, generator.HandlerAnalyzer, "a")
}
