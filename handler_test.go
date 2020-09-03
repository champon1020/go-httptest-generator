package generator_test

import (
	"testing"

	generator "github.com/champon1020/go-httptest-generator"
	"golang.org/x/tools/go/analysis/analysistest"
)

// TestHandlerAnalyzer is a test for HandlerAnalyzer.
func TestHanlderAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, generator.HandlerAnalyzer, "handler/a")
	//analysistest.Run(t, testdata, generator.HandlerAnalyzer, "handler/b")
}
