package ams

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_CreateAssetFile(t *testing.T) {
	m := http.NewServeMux()
	m.HandleFunc("/Files", func(w http.ResponseWriter, r *http.Request) {
		testRequestMethod(t, r, http.MethodPost)
		testAMSHeader(t, r.Header, false)

		var assetFile AssetFile
		if err := json.NewDecoder(r.Body).Decode(&assetFile); err != nil {
			t.Fatal(err)
		}

		if len(assetFile.Name) == 0 {
			t.Error("Name is required")
		}
		if len(assetFile.ParentAssetID) == 0 {
			t.Error("ParentAssetID is required")
		}
		if len(assetFile.MIMEType) == 0 {
			t.Error("MIMEType is required")
		}

		assetFile.ContentFileSize = "0"
		assetFile.Created = formatTime(time.Now())
		assetFile.LastModified = formatTime(time.Now())
		assetFile.ID = "create-asset-file-id"

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(assetFile); err != nil {
			t.Fatal(err)
		}
	})
	s := httptest.NewServer(m)
	defer s.Close()

	client := testClient(t, s.URL)

	assetID := "parent-asset-id"
	name := "test.mp4"
	mime := "video/mp4"
	assetFile, err := client.CreateAssetFile(context.TODO(), assetID, name, mime)
	if err != nil {
		t.Fatal(err)
	}

	tcs := []struct {
		Name     string
		Expected interface{}
		Actual   interface{}
	}{
		{Name: "assetID", Expected: assetID, Actual: assetFile.ParentAssetID},
		{Name: "name", Expected: name, Actual: assetFile.Name},
		{Name: "mime", Expected: mime, Actual: assetFile.MIMEType},
	}

	for _, tc := range tcs {
		if tc.Actual != tc.Expected {
			t.Errorf("unexpected %v. expected: %#v, actual: %#v", tc.Name, tc.Expected, tc.Actual)
		}
	}
}
