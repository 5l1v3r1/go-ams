package ams

import (
	"net/http"
	"strings"

	"testing"
)

type headerMatcher func(string, string) bool

type headerChecker struct {
	t      *testing.T
	header http.Header
}

func (hc *headerChecker) Match(key, expected string, matcher headerMatcher) {
	if actual := hc.header.Get(key); !matcher(actual, expected) {
		hc.t.Errorf("%s: expected %#v, actual %#v", key, expected, actual)
	}
}

func (hc *headerChecker) Assert(key, expected string) {
	hc.Match(key, expected, func(a, b string) bool { return a == b })
}

func (hc *headerChecker) AssertNot(key, unexpected string) {
	hc.Match(key, unexpected, func(a, b string) bool { return a != b })
}

func testAMSHeader(t *testing.T, header http.Header, verbose bool) {
	hc := headerChecker{t, header}
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
	hc := headerChecker{t, header}
	hc.Assert("x-ms-version", StorageAPIVersion)
	hc.AssertNot("Date", "")
}
