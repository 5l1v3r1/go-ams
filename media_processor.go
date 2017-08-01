package ams

import (
	"context"
	"net/http"
)

const (
	mediaProcessorsEndpoint = "MediaProcessors"
)

type MediaProcessor struct {
	ID          string `json:"Id"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	SKU         string `json:"Sku"`
	Vendor      string `json:"Vendor"`
	Version     string `json:"Version"`
}

func (c *Client) GetMediaProcessorsWithContext(ctx context.Context) ([]MediaProcessor, error) {
	req, err := c.newRequest(ctx, http.MethodGet, mediaProcessorsEndpoint, useAMS(c))
	if err != nil {
		return nil, err
	}
	var out struct {
		MediaProcessors []MediaProcessor `json:"value"`
	}
	if err := c.do(req, http.StatusOK, &out); err != nil {
		return nil, err
	}
	return out.MediaProcessors, nil
}
