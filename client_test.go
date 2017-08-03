package ams

import (
	"testing"

	"golang.org/x/oauth2"
)

var (
	dummyTokenSource = oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: "<<dummy access token>>",
		TokenType:   "Bearer",
	})
	dummyURL = "http://example.ams.net/"
)

func TestNewClientWithInvalidURL(t *testing.T) {
	client, err := NewClient("", dummyTokenSource)
	if err == nil {
		t.Errorf("accept invalid url")
	}
	if client != nil {
		t.Errorf("return invalid client")
	}
}

func TestNewClientWithInvalidTokenSource(t *testing.T) {
	client, err := NewClient(dummyURL, nil)
	if err == nil {
		t.Errorf("accept invalid token source")
	}
	if client != nil {
		t.Errorf("return invalid client")
	}
}

func TestNewClient(t *testing.T) {
	client, err := NewClient(dummyURL, dummyTokenSource)
	if err != nil {
		t.Fatal(err)
	}
	if client == nil {
		t.Errorf("return invalid client")
	}
}
