package ams

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

type Locator struct {
	ID                     string `json:"Id"`
	ExpirationDateTime     string `json:"ExpirationDateTime"`
	Type                   int    `json:"Type"`
	Path                   string `json:"Path"`
	BaseURI                string `json:"BaseUri"`
	ContentAccessComponent string `json:"ContentAccessComponent"`
	AccessPolicyID         string `json:"AccessPolicyId"`
	AssetID                string `json:"AssetID"`
	StartTime              string `json:"StartTime"`
	Name                   string `json:"Name"`
}

func (l *Locator) ToUploadURL(name string) (*url.URL, error) {
	uploadURL, err := url.ParseRequestURI(l.Path)
	if err != nil {
		return nil, err
	}
	uploadURL.Path = path.Join(uploadURL.Path, name)
	return uploadURL, nil
}

func (c *Client) CreateLocator(accessPolicyID, assetID, startTime string, locatorType int) (*Locator, error) {
	return c.CreateLocatorWithContext(context.Background(), accessPolicyID, assetID, startTime, locatorType)
}

func (c *Client) CreateLocatorWithContext(ctx context.Context, accessPolicyID, assetID, startTime string, locatorType int) (*Locator, error) {
	params := map[string]interface{}{
		"AccessPolicyId": accessPolicyID,
		"AssetId":        assetID,
		"StartTime":      startTime,
		"Type":           locatorType,
	}
	body, err := encodeParams(params)
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest(ctx, http.MethodPost, "Locators", body)
	if err != nil {
		return nil, err
	}
	var out Locator
	if err := c.do(req, http.StatusCreated, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteLocator(locator *Locator) error {
	return c.DeleteLocatorWithContext(context.Background(), locator)
}

func (c *Client) DeleteLocatorWithContext(ctx context.Context, locator *Locator) error {
	endpoint := fmt.Sprintf("Locators('%s')", url.PathEscape(locator.ID))
	req, err := c.newRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	return c.do(req, http.StatusNoContent, nil)
}
