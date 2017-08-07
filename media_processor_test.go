package ams

import (
	"encoding/json"
	"net/http"
	"testing"

	"context"
	"net/http/httptest"
	"reflect"
	_ "reflect"
)

func TestClient_GetMediaProcessors(t *testing.T) {
	expected := MediaProcessor{
		ID:          "sample-id",
		Name:        "Sample Media Processor",
		Description: "Sample Media Processor, for test",
		SKU:         "Stock Keeping Unit?",
		Vendor:      "Sample Vendor",
		Version:     "1.0.0",
	}
	m := http.NewServeMux()
	m.HandleFunc("/MediaProcessors", func(w http.ResponseWriter, r *http.Request) {
		testRequestMethod(t, r, http.MethodGet)
		testAMSHeader(t, r.Header, false)

		resp := struct {
			Value []MediaProcessor `json:"value"`
		}{
			Value: []MediaProcessor{
				expected,
				expected,
			},
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	})

	ts := httptest.NewServer(m)
	defer ts.Close()

	client, err := NewClient(ts.URL, testTokenSource())
	if err != nil {
		t.Fatal(err)
	}

	mediaProcessors, err := client.GetMediaProcessors(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, mediaProcessors[0]) {
		t.Errorf("unexpected result[0]. expected: %#v, actual: %#v", expected, mediaProcessors[0])
	}

	if !reflect.DeepEqual(expected, mediaProcessors[1]) {
		t.Errorf("unexpected result[1]. expected: %#v, actual: %#v", expected, mediaProcessors[1])
	}
}
