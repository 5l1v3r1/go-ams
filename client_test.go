package ams

import (
	"testing"
	"golang.org/x/oauth2"
)

func TestNewClientWithInvalidURL(t *testing.T) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: "<<dummy access token>>",
		TokenType: "Dummy",
	})
	client, err := NewClient("", ts)
	if err == nil {
		t.Errorf("")
	}
}
