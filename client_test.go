package ams

import (
	"io/ioutil"
	"log"
	"testing"
)

func TestNewClient(t *testing.T) {
	authorizedCliet := testAuthorizedClient()
	dummyURL := "http://example.ams.net/"

	t.Run("withInvalidURL", func(t *testing.T) {
		client, err := NewClient("", authorizedCliet)
		if err == nil {
			t.Errorf("accept invalid url")
		}
		if client != nil {
			t.Errorf("return invalid client")
		}
	})

	t.Run("withInvalidTokenSource", func(t *testing.T) {
		client, err := NewClient(dummyURL, nil)
		if err == nil {
			t.Errorf("accept invalid token source")
		}
		if client != nil {
			t.Errorf("return invalid client")
		}
	})

	t.Run("positiveTest", func(t *testing.T) {
		client, err := NewClient(dummyURL, authorizedCliet)
		if err != nil {
			t.Fatal(err)
		}
		if client == nil {
			t.Errorf("return invalid client")
		}
	})

	t.Run("withUserAgent", func(t *testing.T) {
		expected := "test"
		client, err := NewClient(dummyURL, authorizedCliet, SetUserAgent(expected))
		if err != nil {
			t.Fatal(err)
		}
		if client == nil {
			t.Fatal("return invalid client")
		}
		if client.userAgent != expected {
			t.Errorf("unexpected userAgent. expected: %v, actual: %v", expected, client.userAgent)
		}
	})

	t.Run("withLogger", func(t *testing.T) {
		expected := log.New(ioutil.Discard, "dummy-logger: ", log.LstdFlags)
		client, err := NewClient(dummyURL, authorizedCliet, SetLogger(expected))
		if err != nil {
			t.Fatal(err)
		}
		if client == nil {
			t.Fatal("return invalid client")
		}
		if client.logger != expected {
			t.Errorf("unexpected logger. expected: %#+v, actual: %#+v", expected, client.logger)
		}
	})

	t.Run("withDebug", func(t *testing.T) {
		expected := true
		client, err := NewClient(dummyURL, authorizedCliet, SetDebug(expected))
		if err != nil {
			t.Fatal(err)
		}
		if client == nil {
			t.Fatal("return invalid client")
		}
		if client.debug != expected {
			t.Errorf("unexpected debugFlag. expected: %#+v, actual: %#+v", expected, client.debug)
		}
	})
}
