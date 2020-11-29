package generator

import (
	"html/template"
	"log"
	"os"

	"github.com/champon1020/go-httptest-generator/std"
)

// PreTmplData contains the data about package and import statement.
type PreTmplData struct {
	PkgName     string
	ImportPaths []string
}

var preTmpl = template.Must(template.New("pacakgeAndImport").Parse(`package {{.PkgName}}_test

import (
    "fmt"
    "io/ioutil"
    "net/http/httptest"
    "testing"
    {{range .ImportPaths}}
    "{{.}}"{{end}}
)
`))

// TmplData contains the data for generating test file.
type TmplData struct {
	PkgName      string
	HandlerName  string
	TestFuncName string
	URL          string
	Method       string

	WrapHandlerFunc bool

	// If true, wrap handler with new builtin like `new(someHandler)`.
	// Also call the ServeHTTP function.
	NewHandler bool

	// If true, call the ServeHTTP function.
	InstanceHandler bool
}

// Template for standard httptest.
var tmpl = template.Must(template.New("httptest").Parse(`
// Route "{{.URL}}
// Method "{{.Method}}"
// Handler "{{.PkgName}}.{{.HandlerName}}"
func Test{{.TestFuncName}}(t *testing.T) {
    req := httptest.NewRequest("{{.Method}}", "{{.URL}}", nil)
    resp := httptest.NewRecorder()

    {{if .WrapHandlerFunc}}
    {{if ne .PkgName "main"}}{{.PkgName}}.{{end}}{{.HandlerName}}(resp, req)
    {{else if .NewHandler}}
    handler := new({{if ne .PkgName "main"}}{{.PkgName}}.{{end}}{{.HandlerName}})
    handler.ServeHTTP(resp, req)
    {{else if .InstanceHandler}}
    {{if ne .PkgName "main"}}{{.PkgName}}.{{end}}{{.HandlerName}}.ServeHTTP(resp, req)
    {{else}}
    {{if ne .PkgName "main"}}{{.PkgName}}.{{end}}{{.HandlerName}}(resp, req){{end}}

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        t.Errorf("Failed to read response body %v\n", err)
    }

    fmt.Println(string(data))
}
`))

// GenerateTestFiles generates all test files.
func GenerateTestFiles(hs []*std.Handler) {
	// Sort handlers.
	std.SortHandlers(hs)

	// Insert import and package statement.
	insertPkgImpStmt(hs)

	// Insert test functions.
	for _, h := range hs {
		insertTestFunc(h)
	}
}

// insertPkgImpStmt create pakcage and import statement into test file.
func insertPkgImpStmt(hs []*std.Handler) {
	pkgAndImpTmplData := &PreTmplData{PkgName: hs[0].Pkg.Name}
	fileToTmplMap := make(map[string]*PreTmplData)
	impMap := make(map[string]bool)

	// aggregate
	for _, h := range hs {
		if _, ok := fileToTmplMap[h.File]; !ok {
			// add new
			fileToTmplMap[h.File] = pkgAndImpTmplData
		} else {
			// check whether import path is duplicate or not
			impPath := getImportPath(h.Pkg.Name, h.Pkg.Path)
			if _, ok := impMap[impPath]; !ok {
				// update
				fileToTmplMap[h.File].ImportPaths = append(
					fileToTmplMap[h.File].ImportPaths,
					getImportPath(h.Pkg.Name, h.Pkg.Path),
				)
				impMap[impPath] = true
			}
		}
	}

	// create files
	for fn, data := range fileToTmplMap {
		f, err := os.OpenFile(getTestFileName(fn), os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			/* handler error */
			log.Fatal(err)
		}

		if err := preTmpl.Execute(f, data); err != nil {
			log.Fatal(err)
		}
	}
}

// insertTestFunc create test function to each endpoint.
func insertTestFunc(h *std.Handler) {
	f, err := os.OpenFile(getTestFileName(h.File), os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		/* handle error */
		return
	}

	tmplData := TmplData{
		PkgName:         h.Pkg.Name,
		HandlerName:     h.Name,
		TestFuncName:    getTestFuncName(h.Method, h.URL),
		URL:             h.URL,
		Method:          h.Method,
		WrapHandlerFunc: (h.TypeFlg&1<<std.FuncLitH) != 0 || (h.TypeFlg&1<<std.FuncDeclH) != 0,
		NewHandler:      (h.TypeFlg & 1 << std.NewBuiltinH) != 0,
		InstanceHandler: (h.TypeFlg & 1 << std.InstanceH) != 0,
	}

	if err := tmpl.Execute(f, tmplData); err != nil {
		log.Fatal(err)
	}
}
