package ams

import (
	"context"
	"fmt"
	"html"
	"net/http"

	"github.com/pkg/errors"
	"time"
)

type MetaData struct {
	URI string `json:"uri"`
}

type MediaAsset struct {
	MetaData MetaData `json:"__metadata"`
}

const (
	taskBodyTemplate = "<?xml version=\"1.0\" encoding=\"utf-8\"?><taskBody><inputAsset>JobInputAsset(0)</inputAsset><outputAsset assetName=\"%s\">JobOutputAsset(0)</outputAsset></taskBody>"
)

const (
	JobQueued = iota
	JobScheduled
	JobProcessing
	JobFinished
	JobError
	JobCanceled
	JobCanceling
)

func NewMediaAsset(uri string) MediaAsset {
	return MediaAsset{
		MetaData: MetaData{
			URI: uri,
		},
	}
}

type Task struct {
	Name             string `json:"Name"`
	Configuration    string `json:"Configuration"`
	MediaProcessorID string `json:"MediaProcessorId"`
	TaskBody         string `json:"TaskBody"`
}

type Job struct {
	ID              string  `json:"Id"`
	Name            string  `json:"Name"`
	StartTime       string  `json:"StartTime"`
	EndTime         string  `json:"EndTime"`
	LastModified    string  `json:"LastModified"`
	Priority        int     `json:"Priority"`
	RunningDuration float64 `json:"RunningDuration"`
	State           int     `json:"State"`
}

func (c *Client) EncodeAssetWithContext(ctx context.Context, mediaProcessorID, configuration string, asset *Asset) (*Job, error) {
	destAssetName := fmt.Sprintf("[ENCODED]%s", asset.Name)
	params := map[string]interface{}{
		"Name": fmt.Sprintf("EncodeJob - %s", asset.ID),
		"InputMediaAssets": []MediaAsset{
			NewMediaAsset(c.buildAssetURI(asset)),
		},
		"Tasks": []Task{
			Task{
				Name:             fmt.Sprintf("task-%s", destAssetName),
				Configuration:    configuration,
				MediaProcessorID: mediaProcessorID,
				TaskBody:         buildTaskBody(destAssetName),
			},
		},
	}
	body, err := encodeParams(params)
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest(ctx, http.MethodPost, "Jobs", body)
	req.Header.Set("Content-Type", "application/json;odata=verbose")
	req.Header.Set("Accept", "application/json;odata=verbose")

	var out struct {
		Data Job `json:"d"`
	}
	if err := c.do(req, http.StatusCreated, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

func (c *Client) GetJobWithContext(ctx context.Context, jobID string) (*Job, error) {
	req, err := c.newRequest(ctx, http.MethodGet, fmt.Sprintf("Jobs('%s')", jobID), nil)
	if err != nil {
		return nil, err
	}
	var out Job
	if err := c.do(req, http.StatusOK, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) WaitJobWithContext(ctx context.Context, job *Job) error {
	for {
		current, err := c.GetJobWithContext(ctx, job.ID)
		if err != nil {
			return err
		}

		if current.State == JobError {
			return errors.New("job failed")
		}
		if current.State == JobCanceled {
			return errors.New("job canceled")
		}
		if current.State == JobFinished {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
}

func buildTaskBody(assetName string) string {
	return fmt.Sprintf(taskBodyTemplate, html.EscapeString(assetName))
}
