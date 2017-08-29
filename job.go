package ams

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"path"

	"github.com/pkg/errors"
)

const (
	taskBodyTemplate = `<?xml version="1.0" encoding="utf-8"?><taskBody><inputAsset>JobInputAsset(0)</inputAsset><outputAsset assetName="%s">JobOutputAsset(0)</outputAsset></taskBody>`
)

const (
	jobsEndpoint = "Jobs"
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

type MetaData struct {
	URI string `json:"uri"`
}

type MediaAsset struct {
	MetaData MetaData `json:"__metadata"`
}

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

func (c *Client) EncodeAsset(ctx context.Context, assetID, outputAssetName, mediaProcessorID, configuration string) (*Job, error) {
	jobName := fmt.Sprintf("EncodeJob#%s", assetID)
	assetURI := c.buildAssetURI(assetID)
	taskName := fmt.Sprintf("Task#%s", outputAssetName)
	taskBody := buildTaskBody(outputAssetName)

	params := map[string]interface{}{
		"Name": jobName,
		"InputMediaAssets": []MediaAsset{
			NewMediaAsset(assetURI),
		},
		"Tasks": []Task{
			{
				Name:             taskName,
				Configuration:    configuration,
				MediaProcessorID: mediaProcessorID,
				TaskBody:         taskBody,
			},
		},
	}
	req, err := c.newRequest(ctx, http.MethodPost, jobsEndpoint, withJSON(params), withOData(true))
	if err != nil {
		return nil, errors.Wrap(err, "request build failed")
	}

	c.logger.Printf("[INFO] post encode asset[#%s] job ...", assetID)
	var out struct {
		Data Job `json:"d"`
	}
	if err := c.do(req, http.StatusCreated, &out); err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	c.logger.Printf("[INFO] completed")
	return &out.Data, nil
}

func (c *Client) GetOutputMediaAssets(ctx context.Context, jobID string) ([]Asset, error) {
	endpoint := path.Join(toJobResource(jobID), "OutputMediaAssets")
	req, err := c.newRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "request build failed")
	}
	c.logger.Printf("[INFO] get job[#%s]'s output media assets ...", jobID)
	var out struct {
		Assets []Asset `json:"value"`
	}
	if err := c.do(req, http.StatusOK, &out); err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	c.logger.Printf("[INFO] completed")
	return out.Assets, nil
}

func (c *Client) GetJob(ctx context.Context, jobID string) (*Job, error) {
	endpoint := toJobResource(jobID)
	req, err := c.newRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "request build failed")
	}

	c.logger.Printf("[INFO] get job #%s ...", jobID)
	var out Job
	if err := c.do(req, http.StatusOK, &out); err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	c.logger.Printf("[INFO] completed")
	return &out, nil
}

func buildTaskBody(assetName string) string {
	return fmt.Sprintf(taskBodyTemplate, html.EscapeString(assetName))
}

func toJobResource(jobID string) string {
	return toResource(jobsEndpoint, jobID)
}
