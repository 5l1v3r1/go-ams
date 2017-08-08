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

func testAMSHeader(t *testing.T, header http.Header, verbose bool) {
	hc := checker{t, header}
	hc.Assert("x-ms-version", APIVersion)
	hc.Assert("DataServiceVersion", DataServiceVersion)
	hc.Assert("MaxDataServiceVersion", MaxDataServiceVersion)
	if verbose {
		hc.Assert("Content-Type", "application/json;odata=verbose")
		hc.Assert("Accept", "application/json;odata=verbose")
	} else {
		hc.Assert("Content-Type", "application/json")
		hc.Assert("Accept", "application/json")
	}
	hc.Match("Authorization", "Bearer ", strings.HasPrefix)
}

func testStorageHeader(t *testing.T, header http.Header) {
	hc := checker{t, header}
	hc.Assert("x-ms-version", StorageAPIVersion)
	hc.AssertNot("Date", "")
}
