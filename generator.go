package generator

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

	"github.com/champon1020/go-httptest-generator/handler"
)

type PkgAndImpTmplData struct {
	PkgName     string
	ImportPaths []string
}

var pkgImpTmpl = template.Must(template.New("pacakgeAndImport").Parse(`package {{.PkgName}}_test

import (
    "fmt"
    "io/ioutil"
    "net/http/httptest"
    "testing"
    {{range .ImportPaths}}
    "{{.}}"{{end}}
)
`))

// TestTmplData keeps template data for generating.
type TestTmplData struct {
	PkgName         string
	HandlerName     string
	TestFuncName    string
	URL             string
	Method          string
	WrapHandlerFunc bool
	NewHandler      bool
	InstanceHandler bool
}

// Template for standard httptest.
var stdTmpl = template.Must(template.New("httptest").Parse(`
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

// GenerateAllTests generates all tests.
func GenerateAllTests(contexts []*handler.Context) {
	handler.SortContexts(contexts)
	generatePkgAndImpStmt(contexts)

	for _, ctx := range contexts {
		generateTest(ctx)
	}
}

// Add pakcage and import statement to test file.
func generatePkgAndImpStmt(contexts []*handler.Context) {
	pkgAndImpTmplData := &PkgAndImpTmplData{PkgName: contexts[0].Pkg.Name}
	fileToTmplMap := make(map[string]*PkgAndImpTmplData)
	impMap := make(map[string]bool)

	// aggregate
	for _, ctx := range contexts {
		if _, ok := fileToTmplMap[ctx.File]; !ok {
			// add new
			fileToTmplMap[ctx.File] = pkgAndImpTmplData
		} else {
			// check whether import path is duplicate or not
			impPath := getImportPath(ctx.Pkg.Name, ctx.Pkg.Path)
			if _, ok := impMap[impPath]; !ok {
				// update
				fileToTmplMap[ctx.File].ImportPaths = append(
					fileToTmplMap[ctx.File].ImportPaths,
					getImportPath(ctx.Pkg.Name, ctx.Pkg.Path),
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

		if err := pkgImpTmpl.Execute(f, data); err != nil {
			log.Fatal(err)
		}
	}
}

// Generate test to each endpoint.
func generateTest(ctx *handler.Context) {
	f, err := os.OpenFile(getTestFileName(ctx.File), os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		/* handle error */
		return
	}

	testTmplData := TestTmplData{
		PkgName:         ctx.Pkg.Name,
		HandlerName:     ctx.Name,
		TestFuncName:    getTestFuncName(ctx.Method, ctx.URL),
		URL:             ctx.URL,
		Method:          ctx.Method,
		WrapHandlerFunc: ctx.IsFuncLit || ctx.IsFuncDecl,
		NewHandler:      ctx.IsNew,
		InstanceHandler: ctx.IsInstance,
	}
	if err := stdTmpl.Execute(f, testTmplData); err != nil {
		log.Fatal(err)
	}
}

// Get test function name.
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

// Get test file name.
func getTestFileName(fn string) string {
	s := strings.Split(fn, ".go")
	return fmt.Sprintf("%s_test.go", s[0])
}

// Get modified import path.
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
