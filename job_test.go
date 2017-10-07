package ams

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

type jobRequest struct {
	Name             string
	InputMediaAssets []MediaAsset
	Tasks            []Task
}

func verifyJobRequest(t *testing.T, r io.Reader) *jobRequest {
	var jobRequest jobRequest
	if err := json.NewDecoder(r).Decode(&jobRequest); err != nil {
		t.Fatal(err)
	}

	if len(jobRequest.Name) == 0 {
		t.Error("Name is required")
	}
	if len(jobRequest.InputMediaAssets) == 0 {
		t.Error("InputMediaAssets is required")
	}
	if len(jobRequest.Tasks) == 0 {
		t.Error("Tasks is required")
	}
	return &jobRequest
}

func TestClient_AddEncodeJob(t *testing.T) {
	assetID := "sample-id"
	outputAssetName := "__sample-output-asset-name__"
	mediaProcessorID := "sample-media-processor-id"

	var client *Client

	m := http.NewServeMux()
	m.HandleFunc("/Jobs", func(w http.ResponseWriter, r *http.Request) {
		testRequestMethod(t, r, http.MethodPost)
		testAMSHeader(t, r.Header, true)

		actual := verifyJobRequest(t, r.Body)
		for _, ma := range actual.InputMediaAssets {
			if assetURI := client.buildAssetURI(assetID); ma.MetaData.URI != assetURI {
				t.Errorf("unexpected AssetURI. expected: %v, actual: %v", assetURI, ma.MetaData.URI)
			}
		}
		for _, ta := range actual.Tasks {
			if ta.Configuration != "Adaptive Streaming" {
				t.Errorf("unexpected configuration. expected: %v, actual: %v", "Adaptive Streaming", ta.Configuration)
			}
		}
		w.WriteHeader(http.StatusCreated)
		rawJob := `{"Id":"sample-job-id","Name":"sample-job-name","StartTime":"2017-08-10T02:52:53Z","EndTime":"","LastModified":"2017-08-10T02:52:53Z","Priority":1,"RunningDuration":0.0,"State":0}`
		fmt.Fprint(w, rawJob)
	})
	s := httptest.NewServer(m)
	defer s.Close()

	client = testClient(t, s.URL)

	job, err := client.AddEncodeJob(context.TODO(), assetID, mediaProcessorID, outputAssetName)
	if err != nil {
		t.Error(err)
	}
	if job == nil {
		t.Error("return nil job")
	}
}

func TestClient_GetOutputMediaAssets(t *testing.T) {
	jobID := "sample-job-id"
	expected := []Asset{
		{
			ID:                 "encode-result-asset-id",
			State:              StateInitialized,
			Created:            formatTime(time.Now()),
			LastModified:       formatTime(time.Now()),
			Name:               "EncodeResult",
			Options:            OptionNone,
			FormatOption:       FormatOptionAdaptiveStreaming,
			URI:                "https://fake.url/Asset('encode-result-asset-id')",
			StorageAccountName: "fake-storage-account",
		},
	}
	body := struct {
		Assets []Asset `json:"value"`
	}{
		Assets: expected,
	}

	m := http.NewServeMux()
	m.HandleFunc(fmt.Sprintf("/Jobs('%v')/OutputMediaAssets", jobID),
		testJSONHandler(t, http.MethodGet, false, http.StatusOK, body),
	)
	s := httptest.NewServer(m)
	defer s.Close()

	client := testClient(t, s.URL)
	actual, err := client.GetOutputMediaAssets(context.TODO(), jobID)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("unexpected output media asset. expected: %#v, actual: %#v", expected, actual)
	}
}

func TestClient_GetJob(t *testing.T) {
	expected := &Job{
		ID:              "sample-job-id",
		Name:            "Sample Job",
		StartTime:       formatTime(time.Now()),
		EndTime:         formatTime(time.Now()),
		LastModified:    formatTime(time.Now()),
		Priority:        1,
		RunningDuration: 100,
		State:           StateInitialized,
	}

	m := http.NewServeMux()
	m.HandleFunc(fmt.Sprintf("/Jobs('%v')", expected.ID),
		testJSONHandler(t, http.MethodGet, false, http.StatusOK, expected),
	)
	s := httptest.NewServer(m)
	defer s.Close()

	client := testClient(t, s.URL)
	actual, err := client.GetJob(context.TODO(), expected.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("unexpected job. expected: %#v, actual: %#v", expected, actual)
	}
}
