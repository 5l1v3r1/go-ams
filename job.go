package ams

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type MetaData struct {
	URI string `json:"uri"`
}

type MediaAsset struct {
	MetaData MetaData `json:"__metadata"`
}

func NewMediaAsset(asset *Asset) MediaAsset {
	return MediaAsset{
		MetaData: MetaData{
			URI: fmt.Sprint("https://media.windows.net/api/Assets('%s')", url.PathEscape(asset.ID)),
		},
	}
}

type Task struct {
	Configuration    string `json:"Configuration"`
	MediaProcessorID string `json:"MediaProcessorId"`
}

func (c *Client) EncodeAsset(asset *Asset) error {
	return c.EncodeAssetWithContext(context.Background(), asset)
}

func (c *Client) EncodeAssetWithContext(ctx context.Context, asset *Asset) error {
	params := map[string]interface{}{
		"Name": "EncodeJob",
		"InputMediaAssets": []MediaAsset{
			NewMediaAsset(asset),
		},
		"Tasks": []Task{},
	}
	body, err := encodeParams(params)
	if err != nil {
		return err
	}
	req, err := c.newRequest(ctx, http.MethodPost, "Jobs", body)
	req.Header.Set("Content-Type", "application/json;odata=verbose")
	req.Header.Set("Accept", "application/json;odata=verbose")

	// expected: http.StatusCreated

	return nil
}
