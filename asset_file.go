package ams

import (
	"context"
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

func (c *Client) CreateAssetFile(ctx context.Context, assetID, name, mimeType string) (*AssetFile, error) {
	params := map[string]interface{}{
		"IsEncrypted":   false,
		"IsPrimary":     false,
		"MimeType":      mimeType,
		"Name":          name,
		"ParentAssetId": assetID,
	}
	req, err := c.newRequest(ctx, http.MethodPost, filesEndpoint, withJSON(params))
	if err != nil {
		return nil, errors.Wrap(err, "request build failed")
	}
	var out AssetFile
	if err := c.do(req, http.StatusCreated, &out); err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	return &out, nil
}

func (c *Client) UpdateAssetFile(ctx context.Context, assetFile *AssetFile) error {
	endpoint := toFileResource(assetFile.ID)
	req, err := c.newRequest(ctx, "MERGE", endpoint, withJSON(assetFile))
	if err != nil {
		return errors.Wrap(err, "request build failed")
	}

	if err := c.do(req, http.StatusNoContent, nil); err != nil {
		return errors.Wrap(err, "request failed")
	}

	return nil
}

func toFileResource(assetFileID string) string {
	return toResource(filesEndpoint, assetFileID)
}
