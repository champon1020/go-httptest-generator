package sample_test

import (
    "fmt"
    "io/ioutil"
    "net/http/httptest"
    "testing"
    
    "github.com/champon1020/go-httptest-generator/sample"
)

// Route "/handlerFunc1
// Method "POST"
// Handler "sample.H"
func TestPostHandlerfunc1(t *testing.T) {
    req := httptest.NewRequest("POST", "/handlerFunc1", nil)
    resp := httptest.NewRecorder()

    
    sample.H(resp, req)

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        t.Errorf("Failed to read response body %v\n", err)
    }

    fmt.Println(string(data))
}

// Route "/handlerFunc4
// Method "POST"
// Handler "sample.Index"
func TestPostHandlerfunc4(t *testing.T) {
    req := httptest.NewRequest("POST", "/handlerFunc4", nil)
    resp := httptest.NewRecorder()

    
    sample.Index(resp, req)
    

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        t.Errorf("Failed to read response body %v\n", err)
    }

    fmt.Println(string(data))
}

// Route "/handlerFunc6
// Method "POST"
// Handler "sample.IndexVar"
func TestPostHandlerfunc6(t *testing.T) {
    req := httptest.NewRequest("POST", "/handlerFunc6", nil)
    resp := httptest.NewRecorder()

    
    sample.IndexVar(resp, req)
    

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        t.Errorf("Failed to read response body %v\n", err)
    }

    fmt.Println(string(data))
}

// Route "/handlerFunc7
// Method "POST"
// Handler "sample.Index"
func TestPostHandlerfunc7(t *testing.T) {
    req := httptest.NewRequest("POST", "/handlerFunc7", nil)
    resp := httptest.NewRecorder()

    
    sample.Index(resp, req)
    

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        t.Errorf("Failed to read response body %v\n", err)
    }

    fmt.Println(string(data))
}
