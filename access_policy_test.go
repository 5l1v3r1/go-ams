package ams

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
)

type Matcher func(string, string) bool

type HeaderChecker struct {
	h   http.Header
	err error
}

func NewHeaderChecker(h http.Header) *HeaderChecker {
	return &HeaderChecker{h: h, err: nil}
}

func (hc *HeaderChecker) Match(key, expected string, matcher Matcher) {
	if hc.err != nil {
		return
	}
	hc.err = assertHeader(hc.h, key, expected, matcher)
}

func (hc *HeaderChecker) Assert(key, expected string) {
	hc.Match(key, expected, func(a, b string) bool { return a == b })
}

func (hc *HeaderChecker) Err() error {
	return hc.err
}

func assertHeader(h http.Header, key, expected string, matcher func(string, string) bool) error {
	if actual := h.Get(key); !matcher(actual, expected) {
		return errors.Errorf("%s: expected %#v, actual %#v", key, expected, actual)
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
	hc.Match("Authorization", "Bearer ", strings.HasPrefix)
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
		var params struct {
			Name              string
			DurationInMinutes float64
			Permissions       int

			ID           string `json:"Id"`
			Created      string
			LastModified string
			Metadata     string `json:"odata.metadata"`
		}
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatal(err)
		}
		if len(params.Name) == 0 {
			t.Errorf("Name is required")
		}
		if params.DurationInMinutes <= 0 {
			t.Errorf("DurationInMinutes must be greater than 0")
		}
		if params.Permissions < 0 || 15 < params.Permissions {
			t.Errorf("invalid Permissions")
		}
		params.Metadata = "https://dummy.url"
		params.ID = "nb:pid:UUID:sample-id"
		params.Created = time.Now().UTC().Format(time.RFC3339)
		params.LastModified = time.Now().UTC().Format(time.RFC3339)

		resp, err := json.Marshal(params)
		if err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write(resp)
	})

	s := httptest.NewServer(m)
	defer s.Close()

	client, err := NewClient(s.URL, newDummyTokenSource())
	if err != nil {
		t.Fatal(err)
	}
	ap, err := client.CreateAccessPolicy(context.TODO(), "test", 400, PermissionRead)
	if err != nil {
		t.Fatal(err)
	}
	if ap.Name != "test" {
		t.Errorf("response name unexpected")
	}
	if ap.Permissions != PermissionRead {
		t.Errorf("response permissions unexpected")
	}
	if ap.DurationInMinutes != 400 {
		t.Errorf("response durationInMinutes unexpected")
	}
}

func TestClient_DeleteAccessPolicy(t *testing.T) {
	m := http.NewServeMux()
	m.HandleFunc("/AccessPolicies('nb:pid:UUID:sample-id')", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("method must be DELETE, actual %s", r.Method)
		}
		if err := validateAMSHeader(false, r.Header); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	s := httptest.NewServer(m)
	defer s.Close()

	client, err := NewClient(s.URL, newDummyTokenSource())
	if err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteAccessPolicy(context.TODO(), "nb:pid:UUID:sample-id"); err != nil {
		t.Error(err)
	}
}
