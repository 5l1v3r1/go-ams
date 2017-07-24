package ams

import (
	"context"
	"fmt"
	"net/http"
)

type AssetOption int

const (
	OptionNone                        = 0
	OptionStorageEncrypted            = 1
	OptionCommonEncryptionProtected   = 2
	OptionEnvelopeEncryptionProtected = 4
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
	return fmt.Sprintf("Assets('%s')", a.ID)
}

func (c *Client) GetAsset(assetID string) (*Asset, error) {
	return c.GetAssetWithContext(context.Background(), assetID)
}

func (c *Client) GetAssetWithContext(ctx context.Context, assetID string) (*Asset, error) {
	req, err := c.newRequest(ctx, http.MethodGet, fmt.Sprintf("Assets('%s')", assetID), nil)
	if err != nil {
		return nil, err
	}

	var out Asset
	if err := c.do(req, http.StatusOK, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetAssets() ([]Asset, error) {
	return c.GetAssetsWithContext(context.Background())
}

func (c *Client) GetAssetsWithContext(ctx context.Context) ([]Asset, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "Assets", nil)
	if err != nil {
		return nil, err
	}
	var out struct {
		Assets []Asset `json:"value"`
	}
	if err := c.do(req, http.StatusOK, &out); err != nil {
		return nil, err
	}
	return out.Assets, nil
}

func (c *Client) CreateAsset(name string) (*Asset, error) {
	return c.CreateAssetWithContext(context.Background(), name)
}

func (c *Client) CreateAssetWithContext(ctx context.Context, name string) (*Asset, error) {
	params := map[string]interface{}{
		"Name": name,
	}
	body, err := encodeParams(params)
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest(ctx, http.MethodPost, "Assets", body)
	if err != nil {
		return nil, err
	}
	var out Asset
	if err := c.do(req, http.StatusCreated, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) buildAssetURI(asset *Asset) string {
	return c.buildURI(asset.toResource())
}
