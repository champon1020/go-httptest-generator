package generator

import (
	"html/template"
	"log"
	"os"
)

// Template function for httptest.
var tmpl = template.Must(template.New("httptest").Parse(`func TestHandler(t *testing.T) {
    w := httptest.NewRecorder()
    req := httptest.NewRequest("{{.Method}}", "{{.URL}}", nil)
    {{.Pkg}}.{{.Name}}(w, req)
}
`))

// HandlerInfo includes information of handler.
type HandlerInfo struct {
	Pkg    string // Included pacakge name
	File   string // Included file name
	Name   string // Handler name
	URL    string // Endpoint url
	Method string // Request method
}

// NewHandlerInfo initializes HandlerInfo.
func NewHandlerInfo(pkg string) *HandlerInfo {
	handlerInfo := HandlerInfo{}
	handlerInfo.Method = "GET"
	handlerInfo.Pkg = pkg
	return &handlerInfo
}

// GenerateTest generates httptest files.
func GenerateTest(handlerInfo *HandlerInfo) {
	if err := tmpl.Execute(os.Stdout, handlerInfo); err != nil {
		log.Fatal(err)
	}
}
