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

func TestClient_EncodeAsset(t *testing.T) {
	assetID := "sample-id"
	outputAssetName := "__sample-output-asset-name__"
	taskBody := buildTaskBody(outputAssetName)
	mediaProcessorID := "sample-media-processor-id"
	configuration := "Adaptive Streaming"

	var client *Client

	m := http.NewServeMux()
	m.HandleFunc("/Jobs", func(w http.ResponseWriter, r *http.Request) {
		testRequestMethod(t, r, http.MethodPost)
		testAMSHeader(t, r.Header, true)

		var actual struct {
			Name             string
			InputMediaAssets []MediaAsset
			Tasks            []Task
		}
		if err := json.NewDecoder(r.Body).Decode(&actual); err != nil {
			t.Fatal(err)
		}

		if len(actual.Name) == 0 {
			t.Error("Name is required")
		}
		if len(actual.InputMediaAssets) == 0 {
			t.Error("InputMediaAssets is required")
		}
		for _, ma := range actual.InputMediaAssets {
			if assetURI := client.buildAssetURI(assetID); ma.MetaData.URI != assetURI {
				t.Errorf("unexpected AssetURI. expected: %v, actual: %v", assetURI, ma.MetaData.URI)
			}
		}
		if len(actual.Tasks) == 0 {
			t.Error("Tasks is required")
		}
		for _, task := range actual.Tasks {
			if task.Configuration != configuration {
				t.Errorf("unexpected Configuration. expected: %v, actual: %v", configuration, task.Configuration)
			}
			if task.MediaProcessorID != mediaProcessorID {
				t.Errorf("unexpected MediaProcessorID. expected: %v, actual: %v", mediaProcessorID, task.MediaProcessorID)
			}
			if task.TaskBody != taskBody {
				t.Errorf("unexpected TaskBody. expected: %v, actual: %v", taskBody, task.TaskBody)
			}
		}

		w.WriteHeader(http.StatusCreated)
		rawJob := `{"Id":"sample-job-id","Name":"sample-job-name","StartTime":"2017-08-10T02:52:53Z","EndTime":"","LastModified":"2017-08-10T02:52:53Z","Priority":1,"RunningDuration":0.0,"State":0}`
		fmt.Fprint(w, rawJob)
	})
	s := httptest.NewServer(m)
	client = testClient(t, s.URL)

	job, err := client.EncodeAsset(context.TODO(), assetID, outputAssetName, mediaProcessorID, configuration)
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

	client := testClient(t, s.URL)
	actual, err := client.GetOutputMediaAssets(context.TODO(), jobID)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("unexpected output media asset. expected: %#v, actual: %#v", expected, actual)
	}
}
