package generator

import (
	"fmt"
	"strings"
)

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
