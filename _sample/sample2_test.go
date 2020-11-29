package sample_test

import (
    "fmt"
    "io/ioutil"
    "net/http/httptest"
    "testing"
    
    "github.com/champon1020/go-httptest-generator/sample"
)

// Route "/handleFunc4
// Method "POST"
// Handler "sample.IndexVar2"
func TestPostHandlefunc4(t *testing.T) {
    req := httptest.NewRequest("POST", "/handleFunc4", nil)
    resp := httptest.NewRecorder()

    
    sample.IndexVar2(resp, req)
    

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        t.Errorf("Failed to read response body %v\n", err)
    }

    fmt.Println(string(data))
}

// Route "/handleFunc3
// Method "POST"
// Handler "sample.IndexVar"
func TestPostHandlefunc3(t *testing.T) {
    req := httptest.NewRequest("POST", "/handleFunc3", nil)
    resp := httptest.NewRecorder()

    
    sample.IndexVar(resp, req)
    

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        t.Errorf("Failed to read response body %v\n", err)
    }

    fmt.Println(string(data))
}

// Route "/handleFunc1
// Method "POST"
// Handler "sample.Index"
func TestPostHandlefunc1(t *testing.T) {
    req := httptest.NewRequest("POST", "/handleFunc1", nil)
    resp := httptest.NewRecorder()

    
    sample.Index(resp, req)
    

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        t.Errorf("Failed to read response body %v\n", err)
    }

    fmt.Println(string(data))
}

// Route "/handleFunc5
// Method "POST"
// Handler "sample.Index"
func TestPostHandlefunc5(t *testing.T) {
    req := httptest.NewRequest("POST", "/handleFunc5", nil)
    resp := httptest.NewRecorder()

    
    sample.Index(resp, req)
    

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        t.Errorf("Failed to read response body %v\n", err)
    }

    fmt.Println(string(data))
}

// Route "/handleFunc7
// Method "POST"
// Handler "sample.Index"
func TestPostHandlefunc7(t *testing.T) {
    req := httptest.NewRequest("POST", "/handleFunc7", nil)
    resp := httptest.NewRecorder()

    
    sample.Index(resp, req)
    

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        t.Errorf("Failed to read response body %v\n", err)
    }

    fmt.Println(string(data))
}
