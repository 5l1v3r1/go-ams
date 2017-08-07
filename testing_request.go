package ams

import (
	"net/http"
	"testing"
)

func testRequestMethod(t *testing.T, r *http.Request, expected string) {
	if r.Method != expected {
		t.Fatalf("method must be %v, actual %v", expected, r.Method)
	}
}
