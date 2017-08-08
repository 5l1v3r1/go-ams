package ams

import (
	"net/http"
	"testing"

	"context"
	"net/http/httptest"
	"reflect"
	_ "reflect"
)

func TestClient_GetMediaProcessors(t *testing.T) {
	expected := []MediaProcessor{
		{
			ID:          "sample-id1",
			Name:        "Sample Media Processor",
			Description: "Sample Media Processor, for test",
			SKU:         "Stock Keeping Unit?",
			Vendor:      "Sample Vendor",
			Version:     "1.0.0",
		},
		{
			ID:          "sample-id2",
			Name:        "Standard Media Processor",
			Description: "Standard Media Processor, for test",
			SKU:         "Stock Keeping Unit??",
			Vendor:      "Sample Inc.",
			Version:     "1.1.0",
		},
		{
			ID:          "sample-id3",
			Name:        "Simple Media Processor",
			Description: "Simple Media Processor, for test",
			SKU:         "Stock Keeping Unit???",
			Vendor:      "Simple Vendor",
			Version:     "2.0.0",
		},
	}
	m := http.NewServeMux()
	m.HandleFunc("/MediaProcessors",
		testJSONHandler(t, http.MethodGet, false, http.StatusOK, testWrapValue(expected)),
	)

	s := httptest.NewServer(m)
	defer s.Close()

	client, err := NewClient(s.URL, testTokenSource())
	if err != nil {
		t.Fatal(err)
	}

	actual, err := client.GetMediaProcessors(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("unexpected media processors. expected: %#v, actual: %#v", expected, actual)
	}
}
