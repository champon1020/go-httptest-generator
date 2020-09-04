package sample_test

import (
    "fmt"
    "io/ioutil"
    "net/http/httptest"
    "testing"
    
    "github.com/champon1020/go-httptest-generator/sample"
)

// Route "/handle1
// Method "POST"
// Handler "sample.AnyHandler"
func TestPostHandle1(t *testing.T) {
    req := httptest.NewRequest("POST", "/handle1", nil)
    resp := httptest.NewRecorder()

    
    handler := new(sample.AnyHandler)
    handler.ServeHTTP(resp, req)
    

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        t.Errorf("Failed to read response body %v\n", err)
    }

    fmt.Println(string(data))
}

// Route "/handle3
// Method "POST"
// Handler "sample.A"
func TestPostHandle3(t *testing.T) {
    req := httptest.NewRequest("POST", "/handle3", nil)
    resp := httptest.NewRecorder()

    
    sample.A.ServeHTTP(resp, req)
    

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        t.Errorf("Failed to read response body %v\n", err)
    }

    fmt.Println(string(data))
}

// Route "/handle4
// Method "POST"
// Handler "sample.A2"
func TestPostHandle4(t *testing.T) {
    req := httptest.NewRequest("POST", "/handle4", nil)
    resp := httptest.NewRecorder()

    
    sample.A2.ServeHTTP(resp, req)
    

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        t.Errorf("Failed to read response body %v\n", err)
    }

    fmt.Println(string(data))
}

// Route "/handle6
// Method "POST"
// Handler "sample.AnyHandler"
func TestPostHandle6(t *testing.T) {
    req := httptest.NewRequest("POST", "/handle6", nil)
    resp := httptest.NewRecorder()

    
    handler := new(sample.AnyHandler)
    handler.ServeHTTP(resp, req)
    

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        t.Errorf("Failed to read response body %v\n", err)
    }

    fmt.Println(string(data))
}
