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
		testAMSHeader(t, r.Header, false)

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
	client := testClient(t, s.URL)
	if err := client.DeleteLocator(context.TODO(), locatorID); err != nil {
		t.Error(err)
	}
}