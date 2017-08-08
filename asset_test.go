package ams

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
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

func TestClient_CreateAsset(t *testing.T) {
	m := http.NewServeMux()
	m.HandleFunc("/Assets", func(w http.ResponseWriter, r *http.Request) {
		testRequestMethod(t, r, http.MethodPost)
		testAMSHeader(t, r.Header, false)

		var asset Asset
		if err := json.NewDecoder(r.Body).Decode(&asset); err != nil {
			t.Fatal(err)
		}
		if len(asset.Name) == 0 {
			t.Fatal("Name is required")
		}
		asset.ID = "created-id"
		asset.Created = formatTime(time.Now())
		asset.LastModified = formatTime(time.Now())
		asset.State = StateInitialized
		asset.FormatOption = FormatOptionNoFormat
		asset.StorageAccountName = "sampleStorage"
		asset.URI = "http://your.asset.uri"

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(asset); err != nil {
			t.Fatal(err)
		}
	})

	s := httptest.NewServer(m)
	defer s.Close()

	client, err := NewClient(s.URL, testTokenSource())
	if err != nil {
		t.Fatal(err)
	}

	name := "sample"
	asset, err := client.CreateAsset(context.TODO(), name)
	if err != nil {
		t.Error(err)
	}
	if asset.Name != name {
		t.Errorf("unexpected Name. expected: %#v, actual: %#v", name, asset.Name)
	}
}

func TestClient_GetAssetFiles(t *testing.T) {
	assetID := "test-asset-id"
	expected := []AssetFile{
		{
			ID:              "asset-file-1",
			Name:            "sample1",
			ContentFileSize: "0",
			ParentAssetID:   assetID,
			IsPrimary:       false,
			LastModified:    formatTime(time.Now()),
			Created:         formatTime(time.Now()),
			MIMEType:        "text/plain",
			ContentChecksum: "",
		},
		{
			ID:              "asset-file-2",
			Name:            "sample2",
			ContentFileSize: "100000000000000000",
			ParentAssetID:   assetID,
			IsPrimary:       true,
			LastModified:    formatTime(time.Now()),
			Created:         formatTime(time.Now()),
			MIMEType:        "vide/mp4",
			ContentChecksum: "",
		},
	}
	m := http.NewServeMux()
	m.HandleFunc(fmt.Sprintf("/Assets('%v')/Files", assetID),
		testJSONHandler(t, http.MethodGet, false, http.StatusOK, testWrapValue(expected)),
	)
	s := httptest.NewServer(m)
	defer s.Close()

	client, err := NewClient(s.URL, testTokenSource())
	if err != nil {
		t.Fatal(err)
	}

	actual, err := client.GetAssetFiles(context.TODO(), assetID)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("unexpected asset files. expected: %#v, actual: %#v", expected, actual)
	}
}
