package ams

import (
	"net/http"
	"strings"
	"testing"

	"context"
	"net/http/httptest"

	"github.com/pkg/errors"
)

type HeaderChecker struct {
	h   http.Header
	err error
}

func NewHeaderChecker(h http.Header) *HeaderChecker {
	return &HeaderChecker{h: h, err: nil}
}

func (hc *HeaderChecker) Assert(key, expected string) {
	if hc.err != nil {
		return
	}
	hc.err = assertHeader(hc.h, key, expected)
}

func (hc *HeaderChecker) AssertPrefix(key, expected string) {
	if hc.err != nil {
		return
	}
	hc.err = assertHeaderPrefix(hc.h, key, expected)
}

func (hc *HeaderChecker) Err() error {
	return hc.err
}

func assertHeader(h http.Header, key, expected string) error {
	if actual := h.Get(key); actual != expected {
		return errors.Errorf("%s: expected %#v, actual %#v", key, expected, actual)
	}
	return nil
}

func assertHeaderPrefix(h http.Header, key, expected string) error {
	if actual := h.Get(key); !strings.HasPrefix(actual, expected) {
		return errors.Errorf("%s: expected %#v*, actual %#v", key, expected, actual)
	}
	return nil
}

func validateAMSHeader(verbose bool, h http.Header) error {
	hc := NewHeaderChecker(h)
	hc.Assert("x-ms-version", apiVersion)
	hc.Assert("DataServiceVersion", "3.0")
	hc.Assert("MaxDataServiceVersion", "3.0")
	if verbose {
		hc.Assert("Content-Type", "application/json;odata=verbose")
		hc.Assert("Accept", "application/json;odata=verbose")
	} else {
		hc.Assert("Content-Type", "application/json")
		hc.Assert("Accept", "application/json")
	}
	hc.AssertPrefix("Authorization", "Bearer ")
	return hc.Err()
}

func TestClient_CreateAccessPolicy(t *testing.T) {
	m := http.NewServeMux()
	m.HandleFunc("/AccessPolicies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method must be POST, actual %s", r.Method)
		}
		if err := validateAMSHeader(false, r.Header); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusCreated)
		
	})
	s := httptest.NewServer(m)
	defer s.Close()

	client, err := NewClient(s.URL, dummyTokenSource)
	if err != nil {
		t.Fatal(err)
	}
	ap, err := client.CreateAccessPolicy(context.TODO(), "test", 400, PermissionRead)
	if err != nil {
		t.Fatal(err)
	}
	if ap.Name != "test" {
		t.Fatal("name must be 'test'")
	}
}

func TestClient_DeleteAccessPolicy(t *testing.T) {

}
