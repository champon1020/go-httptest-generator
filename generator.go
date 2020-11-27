package generator

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

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
		WrapHandlerFunc: h.IsFuncLit || h.IsFuncDecl,
		NewHandler:      h.IsNew,
		InstanceHandler: h.IsInstance,
	}

	if err := tmpl.Execute(f, tmplData); err != nil {
		log.Fatal(err)
	}
}

// getTestFuncName generate test function name from request method and endpoint url.
func getTestFuncName(method string, URL string) string {
	var url string
	for _, s := range strings.Split(URL, "/") {
		url += convertToFirstUpperCamel(s)
	}
	return fmt.Sprintf("%s%s",
		convertToFirstUpperCamel(method),
		convertToFirstUpperCamel(url),
	)
}

// getTestFileName generate test file name.
func getTestFileName(fn string) string {
	s := strings.Split(fn, ".go")
	return fmt.Sprintf("%s_test.go", s[0])
}

// getImportPath generate modified import path.
func getImportPath(pkgName string, pkgPath string) string {
	if pkgName == tailOfImportPath(pkgPath) || pkgName == "main" {
		return pkgPath
	}
	return fmt.Sprintf(`%s "%s"`, pkgName, pkgPath)
}

// Compute tail of import path.
func tailOfImportPath(path string) string {
	sp := strings.Split(path, "/")
	return sp[len(sp)-1]
}

// Convert str to camel case.
// This function converts also first character to upper case.
func convertToFirstUpperCamel(str string) string {
	if len(str) == 0 {
		return ""
	}
	if len(str) == 1 {
		return strings.ToUpper(str[0:1])
	}
	return fmt.Sprintf("%s%s", strings.ToUpper(str[0:1]), strings.ToLower(str[1:]))
}
