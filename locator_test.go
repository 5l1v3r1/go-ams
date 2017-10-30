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

func TestLocator_ToUploadURL(t *testing.T) {
	t.Run("withInvalidPath", func(t *testing.T) {
		locator := Locator{
			Path: "http://%?a",
		}
		u, err := locator.ToUploadURL("sample.txt")
		if err == nil {
			t.Error("accept invalid path")
		}
		if u != nil {
			t.Error("return invalid url")
		}
	})

	t.Run("positiveCase", func(t *testing.T) {
		locator := Locator{
			Path: "https://fake.url/upload?with=sas_tokens",
		}
		u, err := locator.ToUploadURL("test.mp4")
		if err != nil {
			t.Error(err)
		}
		expected := "https://fake.url/upload/test.mp4?with=sas_tokens"
		actual := u.String()

		if actual != expected {
			t.Errorf("unexpected UploadURL. expected: %v, actual: %v", expected, actual)
		}
	})
}

func TestClient_CreateLocator(t *testing.T) {
	accessPolicyID := "sample-access-policy-id"
	assetID := "sample-asset-id"
	startTime := time.Now()
	locatorType := LocatorSAS

	expected := &Locator{
		ID:                     "sample-locator-id",
		ExpirationDateTime:     formatTime(time.Now()),
		Type:                   locatorType,
		Path:                   "https://fake.url/upload?with=sas_tokens",
		BaseURI:                "https://fake.url",
		ContentAccessComponent: "",
		AccessPolicyID:         accessPolicyID,
		AssetID:                assetID,
		StartTime:              formatTime(startTime),
		Name:                   "Sample Locator",
	}

	m := http.NewServeMux()
	m.HandleFunc("/Locators", func(w http.ResponseWriter, r *http.Request) {
		testRequestMethod(t, r, http.MethodPost)
		testAMSHeader(t, r, false)

		var params struct {
			AccessPolicyID string `json:"AccessPolicyId"`
			AssetID        string `json:"AssetId"`
			StartTime      string `json:"StartTime"`
			Type           int    `json:"Type"`
		}

		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatal(err)
		}

		if params.AccessPolicyID != accessPolicyID {
			t.Errorf("unexpected AccessPolicyId. expected: %v, actual: %v", accessPolicyID, params.AccessPolicyID)
		}
		if params.AssetID != assetID {
			t.Errorf("unexpected AssetId. expected: %v, actual: %v", assetID, params.AssetID)
		}
		if params.StartTime != formatTime(startTime) {
			t.Errorf("unexpected StartTime. expected: %v, actual: %v", formatTime(startTime), params.StartTime)
		}
		if params.Type != locatorType {
			t.Errorf("unexpected Type. expected: %v, actual: %v", locatorType, params.Type)
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(expected); err != nil {
			t.Fatal(err)
		}
	})

	s := httptest.NewServer(m)
	defer s.Close()

	client := testClient(t, s.URL)

	actual, err := client.CreateLocator(context.TODO(), accessPolicyID, assetID, startTime, locatorType)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("unexpected locator. expected: %#v, actual: %#v", expected, actual)
	}
}

func TestClient_DeleteLocator(t *testing.T) {
	locatorID := "delete-locator-id"

	m := http.NewServeMux()
	m.HandleFunc(fmt.Sprintf("/Locators('%v')", locatorID),
		testJSONHandler(t, http.MethodDelete, false, http.StatusNoContent, nil),
	)
	s := httptest.NewServer(m)
	defer s.Close()

	client := testClient(t, s.URL)
	if err := client.DeleteLocator(context.TODO(), locatorID); err != nil {
		t.Error(err)
	}
}

func TestClient_GetLocators(t *testing.T) {
	m := http.NewServeMux()

	expected := []Locator{
		{
			ID:                     "sample-locator-id-1",
			ExpirationDateTime:     formatTime(time.Now()),
			Type:                   LocatorSAS,
			Path:                   "https://fake.url/upload?with=sas_tokens",
			BaseURI:                "https://fake.url",
			ContentAccessComponent: "",
			AccessPolicyID:         "dummy-access-policy-id-1",
			AssetID:                "sample-asset-id-1",
			StartTime:              formatTime(time.Now()),
			Name:                   "Sample Locator 1",
		},
		{
			ID:                     "sample-locator-id-2",
			ExpirationDateTime:     formatTime(time.Now()),
			Type:                   LocatorSAS,
			Path:                   "https://fake.url/upload?with=sas_tokens",
			BaseURI:                "https://fake.url",
			ContentAccessComponent: "",
			AccessPolicyID:         "dummy-access-policy-id-2",
			AssetID:                "sample-asset-id-2",
			StartTime:              formatTime(time.Now()),
			Name:                   "Sample Locator 2",
		},
	}
	resp := struct {
		Value []Locator
	}{
		Value: expected,
	}
	m.HandleFunc("/Locators", testJSONHandler(t, http.MethodGet, false, http.StatusOK, resp))
	s := httptest.NewServer(m)
	defer s.Close()

	client := testClient(t, s.URL)
	actual, err := client.GetLocators(context.TODO())
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("unexpected locators. expected: %#v, actual: %#v", expected, actual)
	}
}

func TestClient_GetLocatorsWithAsset(t *testing.T) {
	m := http.NewServeMux()

	expected := []Locator{
		{
			ID:                     "sample-locator-id-1",
			ExpirationDateTime:     formatTime(time.Now()),
			Type:                   LocatorSAS,
			Path:                   "https://fake.url/upload?with=sas_tokens",
			BaseURI:                "https://fake.url",
			ContentAccessComponent: "",
			AccessPolicyID:         "dummy-access-policy-id-1",
			AssetID:                "sample-asset-id-1",
			StartTime:              formatTime(time.Now()),
			Name:                   "Sample Locator 1",
		},
		{
			ID:                     "sample-locator-id-2",
			ExpirationDateTime:     formatTime(time.Now()),
			Type:                   LocatorSAS,
			Path:                   "https://fake.url/upload?with=sas_tokens",
			BaseURI:                "https://fake.url",
			ContentAccessComponent: "",
			AccessPolicyID:         "dummy-access-policy-id-2",
			AssetID:                "sample-asset-id-1",
			StartTime:              formatTime(time.Now()),
			Name:                   "Sample Locator 2",
		},
	}
	resp := struct {
		Value []Locator
	}{
		Value: expected,
	}
	m.HandleFunc(fmt.Sprintf("/Assets('%v')/Locators", expected[0].AssetID), testJSONHandler(t, http.MethodGet, false, http.StatusOK, resp))
	s := httptest.NewServer(m)
	defer s.Close()

	client := testClient(t, s.URL)

	t.Run("invalidAssetID", func(t *testing.T) {
		actual, err := client.GetLocatorsWithAsset(context.TODO(), "%#")
		if err == nil {
			t.Error("accept invalid assetID")
		}
		if actual != nil {
			t.Errorf("return invalid locators")
		}
	})

	t.Run("positiveCase", func(t *testing.T) {
		actual, err := client.GetLocatorsWithAsset(context.TODO(), expected[0].AssetID)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("unexpected locators. expected: %#v, actual: %#v", expected, actual)
		}
	})
}
