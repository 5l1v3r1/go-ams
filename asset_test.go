package ams

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestClient_GetAsset(t *testing.T) {
	expected := testAsset("sample-id", "Sample")
	m := http.NewServeMux()
	m.HandleFunc(fmt.Sprintf("/Assets('%v')", expected.ID),
		testJSONHandler(t, http.MethodGet, false, http.StatusOK, expected),
	)

	s := httptest.NewServer(m)
	defer s.Close()

	client, err := NewClient(s.URL, testTokenSource())
	if err != nil {
		t.Fatal(err)
	}

	actual, err := client.GetAsset(context.TODO(), expected.ID)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("unexpected asset. expected: %#v, actual: %#v", expected, actual)
	}
}

func TestClient_GetAssets(t *testing.T) {
	expected := []Asset{
		testAsset("asset1", "Sample 1"),
		testAsset("asset2", "Sample 2"),
		testAsset("asset4", "Example 1"),
	}
	m := http.NewServeMux()
	m.HandleFunc("/Assets",
		testJSONHandler(t, http.MethodGet, false, http.StatusOK, testWrapValue(expected)),
	)

	s := httptest.NewServer(m)
	defer s.Close()

	client, err := NewClient(s.URL, testTokenSource())
	if err != nil {
		t.Fatal(err)
	}

	actual, err := client.GetAssets(context.TODO())
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("unexpected assets. expected: %#v, actual: %#v", expected, actual)
	}
}
