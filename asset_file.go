package ams

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
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

func (c *Client) CreateAssetFile(assetID, name, mimeType string) (*AssetFile, error) {
	return c.CreateAssetFileWithContext(context.Background(), assetID, name, mimeType)
}

func (c *Client) CreateAssetFileWithContext(ctx context.Context, assetID, name, mimeType string) (*AssetFile, error) {
	params := map[string]interface{}{
		"IsEncrypted":   "false",
		"IsPrimary":     "false",
		"MimeType":      mimeType,
		"Name":          name,
		"ParentAssetId": assetID,
	}
	body, err := encodeParams(params)
	if err != nil {
		return nil, errors.Wrap(err, "create asset file parameter encode failed")
	}

	req, err := c.newRequest(ctx, http.MethodPost, "Files", body)
	if err != nil {
		return nil, errors.Wrap(err, "create asset file request build failed")
	}
	var out AssetFile
	if err := c.do(req, http.StatusCreated, &out); err != nil {
		return nil, errors.Wrap(err, "create asset file request failed")
	}
	return &out, nil
}

func (c *Client) UpdateAssetFile(assetFile *AssetFile) error {
	return c.UpdateAssetFileWithContext(context.Background(), assetFile)
}

func (c *Client) UpdateAssetFileWithContext(ctx context.Context, assetFile *AssetFile) error {
	endpoint := fmt.Sprintf("Files('%s')", assetFile.ID)
	body, err := json.Marshal(assetFile)
	if err != nil {
		return errors.Wrap(err, "asset file marshal failed")
	}

	req, err := c.newRequest(ctx, "MERGE", endpoint, bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "update asset file request build failed")
	}

	if err := c.do(req, http.StatusNoContent, nil); err != nil {
		return errors.Wrap(err, "update asset file request failed")
	}
}
