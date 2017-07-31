package ams

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

const (
	filesEndpoint = "Files"
)

type AssetFile struct {
	ID              string `json:"Id"`
	Name            string `json:"Name"`
	ContentFileSize string `json:"ContentFileSize"`
	ParentAssetID   string `json:"ParentAssetId"`
	IsPrimary       bool   `json:"IsPrimary"`
	LastModified    string `json:"LastModified"`
	Created         string `json:"Created"`
	MIMEType        string `json:"MimeType"`
	ContentChecksum string `json:"ContentChecksum"`
}

func (a *AssetFile) toResource() string {
	return fmt.Sprintf("%s('%s')", filesEndpoint, a.ID)
}

func (c *Client) CreateAssetFileWithContext(ctx context.Context, assetID, name, mimeType string) (*AssetFile, error) {
	params := map[string]interface{}{
		"IsEncrypted":   false,
		"IsPrimary":     false,
		"MimeType":      mimeType,
		"Name":          name,
		"ParentAssetId": assetID,
	}
	body, err := encodeParams(params)
	if err != nil {
		return nil, errors.Wrap(err, "create asset file parameter encode failed")
	}

	req, err := c.newRequest(ctx, http.MethodPost, filesEndpoint, body)
	if err != nil {
		return nil, errors.Wrap(err, "create asset file request build failed")
	}
	var out AssetFile
	if err := c.do(req, http.StatusCreated, &out); err != nil {
		return nil, errors.Wrap(err, "create asset file request failed")
	}
	return &out, nil
}

func (c *Client) UpdateAssetFileWithContext(ctx context.Context, assetFile *AssetFile) error {
	endpoint := assetFile.toResource()
	body, err := encodeParams(assetFile)

	req, err := c.newRequest(ctx, "MERGE", endpoint, body)
	if err != nil {
		return errors.Wrap(err, "update asset file request build failed")
	}

	if err := c.do(req, http.StatusNoContent, nil); err != nil {
		return errors.Wrap(err, "update asset file request failed")
	}

	return nil
}
