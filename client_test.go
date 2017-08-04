package ams

import (
	"testing"

	"golang.org/x/oauth2"
)

func TestNewClient(t *testing.T) {
	dummyTokenSource := newDummyTokenSource()
	dummyURL := "http://example.ams.net/"

	t.Run("withInvalidURL", func(t *testing.T) {
		client, err := NewClient("", dummyTokenSource)
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

	t.Run("positive testing", func(t *testing.T) {
		client, err := NewClient(dummyURL, dummyTokenSource)
		if err != nil {
			t.Fatal(err)
		}
		if client == nil {
			t.Errorf("return invalid client")
		}
	})
}

func newDummyTokenSource() oauth2.TokenSource {
	dummyTokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: "<<dummy access token>>",
		TokenType:   "Bearer",
	})
	return dummyTokenSource
}
