package ams

import (
	"context"
	"encoding/xml"
	"fmt"
	"path"

	"github.com/pkg/errors"
)

const (
	jobsEndpoint   = "Jobs"
	jobInputAsset  = "JobInputAsset(0)"
	jobOutputAsset = "JobOutputAsset(0)"
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

func (c *Client) AddEncodeJob(ctx context.Context, assetID, outputAssetName, mediaProcessorID, configuration string) (*Job, error) {
	c.logger.Printf("[INFO] post encode asset[#%s] job ...", assetID)

	jobName := fmt.Sprintf("EncodeJob#%s", assetID)
	assetURI := c.buildAssetURI(assetID)
	taskName := fmt.Sprintf("Task#%s", outputAssetName)
	taskBody, err := xml.Marshal(TaskBody{
		InputAsset: AssetTag{Asset: jobInputAsset},
		OutputAsset: AssetTag{
			Asset: jobOutputAsset,
			Name:  outputAssetName,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "taskBody xml marshal failed")
	}

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
				TaskBody:         string(taskBody),
			},
		},
	}
	var out struct {
		Data Job `json:"d"`
	}
	if err := c.post(ctx, jobsEndpoint, params, &out, withOData(true)); err != nil {
		return nil, err
	}

	c.logger.Printf("[INFO] completed")
	return &out.Data, nil
}

func (c *Client) GetOutputMediaAssets(ctx context.Context, jobID string) ([]Asset, error) {
	c.logger.Printf("[INFO] get job[#%s]'s output media assets ...", jobID)

	endpoint := path.Join(toJobResource(jobID), "OutputMediaAssets")
	var out struct {
		Assets []Asset `json:"value"`
	}
	if err := c.get(ctx, endpoint, &out); err != nil {
		return nil, err
	}

	c.logger.Printf("[INFO] completed")
	return out.Assets, nil
}

func (c *Client) GetJob(ctx context.Context, jobID string) (*Job, error) {
	c.logger.Printf("[INFO] get job #%s ...", jobID)

	endpoint := toJobResource(jobID)
	var out Job
	if err := c.get(ctx, endpoint, &out); err != nil {
		return nil, err
	}

	c.logger.Printf("[INFO] completed")
	return &out, nil
}

func toJobResource(jobID string) string {
	return toResource(jobsEndpoint, jobID)
}
