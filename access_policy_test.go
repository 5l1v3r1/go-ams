package ams

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_CreateAccessPolicy(t *testing.T) {
	m := http.NewServeMux()
	m.HandleFunc("/AccessPolicies", func(w http.ResponseWriter, r *http.Request) {
		testRequestMethod(t, r, http.MethodPost)
		testAMSHeader(t, r.Header, false)

		var params struct {
			Name              string
			DurationInMinutes float64
			Permissions       int

			ID           string `json:"Id"`
			Created      string
			LastModified string
			Metadata     string `json:"odata.metadata"`
		}

		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatal(err)
		}

		if len(params.Name) == 0 {
			t.Errorf("Name is required")
		}

		if params.DurationInMinutes <= 0 {
			t.Errorf("DurationInMinutes must be greater than 0")
		}

		if params.Permissions < 0 || 15 < params.Permissions {
			t.Errorf("invalid Permissions")
		}

		params.Metadata = "https://dummy.url"
		params.ID = "nb:pid:UUID:sample-id"
		params.Created = formatTime(time.Now())
		params.LastModified = formatTime(time.Now())

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(params); err != nil {
			t.Fatal(err)
		}
	})

	s := httptest.NewServer(m)
	defer s.Close()

	client := testClient(t, s.URL)

	name := "test"
	durationInMinutes := 400.0
	permissions := PermissionRead
	ap, err := client.CreateAccessPolicy(context.TODO(), name, durationInMinutes, permissions)
	if err != nil {
		t.Fatal(err)
	}
	if ap.Name != name {
		t.Errorf("response name unexpected")
	}
	if ap.Permissions != permissions {
		t.Errorf("response permissions unexpected")
	}
	if ap.DurationInMinutes != durationInMinutes {
		t.Errorf("response durationInMinutes unexpected")
	}
}

func TestClient_DeleteAccessPolicy(t *testing.T) {
	accessPolicyID := "nb:pid:UUID:sample-policy"
	m := http.NewServeMux()
	m.HandleFunc(fmt.Sprintf("/AccessPolicies('%v')", accessPolicyID),
		testJSONHandler(t, http.MethodDelete, false, http.StatusNoContent, nil),
	)

	s := httptest.NewServer(m)
	defer s.Close()

	client := testClient(t, s.URL)
	if err := client.DeleteAccessPolicy(context.TODO(), accessPolicyID); err != nil {
		t.Error(err)
	}
}
