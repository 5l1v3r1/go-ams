package ams

import (
	"net/http"
	"testing"
)

func testRequestMethod(t *testing.T, r *http.Request, expected string) {
	if r.Method != expected {
		t.Fatalf("unexpected method. expected: %v, but got: %v", expected, r.Method)
	}
}
