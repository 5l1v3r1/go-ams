package ams

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"context"

	"golang.org/x/oauth2"
)

func testAuthorizedClient() *http.Client {
	return oauth2.NewClient(context.TODO(),
		oauth2.StaticTokenSource(&oauth2.Token{
			TokenType:   "Bearer",
			AccessToken: "<<DUMMY>>",
		}))
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
		testAMSHeader(t, r, verbose)

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
	client, err := NewClient(urlStr, testAuthorizedClient())
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func testTempFile(t *testing.T) (string, func()) {
	tf, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatal(err)
	}
	tf.Close()

	return tf.Name(), func() { os.Remove(tf.Name()) }
}
