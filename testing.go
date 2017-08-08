package ams

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func testTokenSource() oauth2.TokenSource {
	return oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: "<<DUMMY ACCESS TOKEN>>",
		TokenType:   "Bearer",
	})
}

func testAsset(id, name string) Asset {
	return Asset{
		ID:           id,
		State:        StateInitialized,
		Created:      formatTime(time.Now()),
		LastModified: formatTime(time.Now()),
		Name:         name,
		Options:      OptionNone,
		FormatOption: FormatOptionNoFormat,
	}
}

func testJSONHandler(t *testing.T, method string, verbose bool, statusCode int, resp interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		testRequestMethod(t, r, method)
		testAMSHeader(t, r.Header, verbose)

		w.WriteHeader(statusCode)
		if resp != nil {
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func testWrapValue(value interface{}) interface{} {
	return struct {
		Value interface{} `json:"value"`
	}{
		Value: value,
	}
}

func testClient(t *testing.T, urlStr string) *Client {
	client, err := NewClient(urlStr, testTokenSource())
	if err != nil {
		t.Fatal(err)
	}
	return client
}
