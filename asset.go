package ams

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
)

const (
	assetsEndpoint = "Assets"
)

const (
	OptionStorageEncrypted = 1 << iota
	OptionCommonEncryptionProtected
	OptionEnvelopeEncryptionProtected
	OptionNone = 0
)

type Asset struct {
	ID                 string `json:"Id"`
	State              int    `json:"State"`
	Created            string `json:"Created"`
	LastModified       string `json:"LastModified"`
	Name               string `json:"Name"`
	Options            int    `json:"Options"`
	FormatOption       int    `json:"FormatOption"`
	URI                string `json:"Uri"`
	StorageAccountName string `json:"StorageAccountName"`
}

func (c *Client) GetAsset(ctx context.Context, assetID string) (*Asset, error) {
	endpoint := toAssetResource(assetID)
	req, err := c.newRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "request build failed")
	}

	var out Asset
	if err := c.do(req, http.StatusOK, &out); err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	return &out, nil
}

func (c *Client) GetAssets(ctx context.Context) ([]Asset, error) {
	req, err := c.newRequest(ctx, http.MethodGet, assetsEndpoint)
	if err != nil {
		return nil, errors.Wrap(err, "request build failed")
	}
	var out struct {
		Assets []Asset `json:"value"`
	}
	if err := c.do(req, http.StatusOK, &out); err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	return out.Assets, nil
}

func (c *Client) CreateAsset(ctx context.Context, name string) (*Asset, error) {
	params := map[string]interface{}{
		"Name": name,
	}
	req, err := c.newRequest(ctx, http.MethodPost, assetsEndpoint, withJSON(params))
	if err != nil {
		return nil, errors.Wrap(err, "request build failed")
	}
	var out Asset
	if err := c.do(req, http.StatusCreated, &out); err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	return &out, nil
}

func (c *Client) GetAssetFiles(ctx context.Context, assetID string) ([]AssetFile, error) {
	endpoint := toAssetResource(assetID) + "/Files"
	req, err := c.newRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "request build failed")
	}
	var out struct {
		AssetFiles []AssetFile `json:"value"`
	}
	if err := c.do(req, http.StatusOK, &out); err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	return out.AssetFiles, nil
}

func toAssetResource(assetID string) string {
	return toResource(assetsEndpoint, assetID)
}

func (c *Client) buildAssetURI(assetID string) string {
	return c.buildURI(toResource(assetsEndpoint, assetID))
}
