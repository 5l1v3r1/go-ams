package ams

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	tokenSource := testTokenSource()
	dummyURL := "http://example.ams.net/"

	t.Run("withInvalidURL", func(t *testing.T) {
		client, err := NewClient("", tokenSource)
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
		client, err := NewClient(dummyURL, tokenSource)
		if err != nil {
			t.Fatal(err)
		}
		if client == nil {
			t.Errorf("return invalid client")
		}
	})
}
