package generator

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"sort"
	"strings"
)

type PkgAndImpTmplData struct {
	PkgName     string
	ImportPaths []string
}

var pkgImpTmpl = template.Must(template.New("pacakgeAndImport").Parse(`package {{.PkgName}}_test

import (
    "fmt"
    "ioutil"
    "net/http"
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
}

// Template for standard httptest.
var stdTmpl = template.Must(template.New("httptest").Parse(`
// Route "{{.URL}}
// Method "{{.Method}}"
// Handler "{{.PkgName}}.{{.HandlerName}}"
func Test{{.TestFuncName}}(t *testing.T) {
    req := httptest.NewRequest("{{.Method}}", "{{.URL}}", nil)

    {{if .WrapHandlerFunc}}
    ts := httptest.NewServer(http.HandlerFunc({{if ne .PkgName "main"}}{{.PkgName}}.{{end}}{{.HandlerName}}))
    {{else if .NewHandler}}
    ts := httptest.NewServer(new({{if ne .PkgName "main"}}{{.PkgName}}.{{end}}{{.HandlerName}}))
    {{else}}
    ts := httptest.NewServer({{if ne .PkgName "main"}}{{.PkgName}}.{{end}}{{.HandlerName}})    
    {{end}}
    defer ts.Close()

    c := new(http.Client)
    resp, err := c.Do(req)
    if err != nil {
        t.Errorf("Failed to create client %v\n", err)
    }

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        t.Errorf("Failed to read response body %v\n", err)
    }

    fmt.Println(string(data))
}
`))

// Package includes package name and path.
type Package struct {
	Name string
	Path string
}

// HandlerInfo includes information of handler.
type HandlerInfo struct {
	Pkg    Package // Included pacakge
	File   string  // Included file name
	Name   string  // Handler name
	URL    string  // Endpoint url
	Method string  // Request method

	IsFuncLit  bool
	IsFuncDecl bool
	IsNew      bool
	IsInstance bool
}

// NewHandlerInfo initializes HandlerInfo.
func NewHandlerInfo(pkgName string, pkgPath string) *HandlerInfo {
	handlerInfo := HandlerInfo{}
	handlerInfo.Method = "GET"
	handlerInfo.Pkg.Name = pkgName
	handlerInfo.Pkg.Path = pkgPath
	return &handlerInfo
}

// Sort slice of HanlderInfo.
func sortHandlersInfo(h []*HandlerInfo) {
	sort.Slice(h, func(i, j int) bool {
		if h[i].Pkg.Name == h[j].Pkg.Name {
			return h[i].File < h[j].File
		}
		return h[i].Pkg.Name < h[j].Pkg.Name
	})
}

// GenerateAllTests generates all tests.
func GenerateAllTests(handlersInfo []*HandlerInfo) {
	sortHandlersInfo(handlersInfo)
	generatePkgAndImpStmt(handlersInfo)

	for _, h := range handlersInfo {
		generateTest(h)
	}
}

// Add pakcage and import statement to test file.
func generatePkgAndImpStmt(handlersInfo []*HandlerInfo) {
	pkgAndImpTmplData := &PkgAndImpTmplData{PkgName: handlersInfo[0].Pkg.Name}
	fileToTmplMap := make(map[string]*PkgAndImpTmplData)
	impMap := make(map[string]bool)

	// aggregate
	for _, h := range handlersInfo {
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

		if err := pkgImpTmpl.Execute(f, data); err != nil {
			log.Fatal(err)
		}
	}
}

// Generate test to each endpoint.
func generateTest(handlerInfo *HandlerInfo) {
	f, err := os.OpenFile(getTestFileName(handlerInfo.File), os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		/* handle error */
		return
	}

	testTmplData := TestTmplData{
		PkgName:         handlerInfo.Pkg.Name,
		HandlerName:     handlerInfo.Name,
		TestFuncName:    getTestFuncName(handlerInfo.Method, handlerInfo.URL),
		URL:             handlerInfo.URL,
		Method:          handlerInfo.Method,
		WrapHandlerFunc: handlerInfo.IsFuncLit || handlerInfo.IsFuncDecl,
		NewHandler:      handlerInfo.IsNew,
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
