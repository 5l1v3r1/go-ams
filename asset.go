package ams

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
)

type AssetOption int

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

func (a *Asset) toResource() string {
	return toResource(assetsEndpoint, a.ID)
}

func (c *Client) GetAssetWithContext(ctx context.Context, assetID string) (*Asset, error) {
	endpoint := toResource(assetsEndpoint, assetID)
	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, errors.Wrap(err, "request build failed")
	}

	var out Asset
	if err := c.do(req, http.StatusOK, &out); err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	return &out, nil
}

func (c *Client) GetAssetsWithContext(ctx context.Context) ([]Asset, error) {
	req, err := c.newRequest(ctx, http.MethodGet, assetsEndpoint, nil)
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

func (c *Client) CreateAssetWithContext(ctx context.Context, name string) (*Asset, error) {
	params := map[string]interface{}{
		"Name": name,
	}
	body, err := encodeParams(params)
	if err != nil {
		return nil, errors.Wrap(err, "request parameter encode failed")
	}
	req, err := c.newRequest(ctx, http.MethodPost, assetsEndpoint, body)
	if err != nil {
		return nil, errors.Wrap(err, "request build failed")
	}
	var out Asset
	if err := c.do(req, http.StatusCreated, &out); err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	return &out, nil
}

func (c *Client) GetAssetFilesWithContext(ctx context.Context, asset *Asset) ([]AssetFile, error) {
	endpoint := asset.toResource() + "/Files"
	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
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

func (c *Client) buildAssetURI(asset *Asset) string {
	return c.buildURI(asset.toResource())
}
