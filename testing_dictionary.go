package ams

import (
	"net/http"
	"strings"

	"testing"
)

type dictionary interface {
	Get(string) string
}

type headerMatcher func(string, string) bool

type checker struct {
	t          *testing.T
	dictionary dictionary
}

func (hc *checker) Match(key, expected string, matcher headerMatcher) {
	if actual := hc.dictionary.Get(key); !matcher(actual, expected) {
		hc.t.Errorf("%s: expected %#v, actual %#v", key, expected, actual)
	}
}

func (hc *checker) Assert(key, expected string) {
	hc.Match(key, expected, func(a, b string) bool { return a == b })
}

func (hc *checker) AssertNot(key, unexpected string) {
	hc.Match(key, unexpected, func(a, b string) bool { return a != b })
}

func testAMSHeader(t *testing.T, req *http.Request, verbose bool) {
	hc := checker{t, req.Header}
	hc.Assert("x-ms-version", APIVersion)
	hc.Assert("DataServiceVersion", DataServiceVersion)
	hc.Assert("MaxDataServiceVersion", MaxDataServiceVersion)

	checkContentType := req.Method == http.MethodPost || req.Method == http.MethodPut
	if verbose {
		if checkContentType {
			hc.Assert("Content-Type", "application/json;odata=verbose")
		}
		hc.Assert("Accept", "application/json;odata=verbose")
	} else {
		if checkContentType {
			hc.Assert("Content-Type", "application/json")
		}
		hc.Assert("Accept", "application/json")
	}
	hc.Match("Authorization", "Bearer ", strings.HasPrefix)
}
