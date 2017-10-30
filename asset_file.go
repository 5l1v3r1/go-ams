package ams

import (
	"context"
	"net/http"

	"github.com/orisano/httpc"
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
	c.logger.Printf("[INFO] create asset[#%s] file ...", assetID)

	params := map[string]interface{}{
		"IsEncrypted":   false,
		"IsPrimary":     false,
		"MimeType":      mimeType,
		"Name":          name,
		"ParentAssetId": assetID,
	}
	var out AssetFile
	if err := c.post(ctx, filesEndpoint, params, &out); err != nil {
		return nil, err
	}

	c.logger.Printf("[INFO] completed, new asset[#%s] file[#%s]", assetID, out.ID)
	return &out, nil
}

func (c *Client) UpdateAssetFile(ctx context.Context, assetFile *AssetFile) error {
	endpoint := toFileResource(assetFile.ID)
	req, err := c.newRequest(ctx, "MERGE", endpoint, httpc.WithJSON(assetFile))
	if err != nil {
		return errors.Wrap(err, "request build failed")
	}

	c.logger.Printf("[INFO] update asset[#%s] file[#%s] ...", assetFile.ParentAssetID, assetFile.ID)
	if err := c.do(req, http.StatusNoContent, nil); err != nil {
		return errors.Wrap(err, "request failed")
	}
	c.logger.Printf("[INFO] completed")

	return nil
}

func toFileResource(assetFileID string) string {
	return toResource(filesEndpoint, assetFileID)
}
