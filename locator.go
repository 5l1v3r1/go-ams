package ams

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/pkg/errors"
)

const (
	locatorsEndpoint = "Locators"
)

const (
	LocatorNone = iota
	LocatorSAS
	LocatorOnDemandOrigin
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
		return nil, errors.Wrap(err, "failed to parse url")
	}
	uploadURL.Path = path.Join(uploadURL.Path, name)
	return uploadURL, nil
}

func (c *Client) CreateLocator(ctx context.Context, accessPolicyID, assetID string, startTime time.Time, locatorType int) (*Locator, error) {
	c.logger.Printf("[INFO] create locator ...")

	params := map[string]interface{}{
		"AccessPolicyId": accessPolicyID,
		"AssetId":        assetID,
		"StartTime":      formatTime(startTime),
		"Type":           locatorType,
	}
	var out Locator
	if err := c.post(ctx, locatorsEndpoint, params, &out); err != nil {
		return nil, err
	}

	c.logger.Printf("[INFO] completed, new locator[#%s]", out.ID)
	return &out, nil
}

func (c *Client) DeleteLocator(ctx context.Context, locatorID string) error {
	endpoint := toLocatorResource(locatorID)
	req, err := c.newRequest(ctx, http.MethodDelete, endpoint)
	if err != nil {
		return errors.Wrap(err, "failed to construct request")
	}
	c.logger.Printf("[INFO] delete locator #%s ...", locatorID)
	if err := c.do(req, http.StatusNoContent, nil); err != nil {
		return errors.Wrap(err, "failed to request")
	}
	c.logger.Printf("[INFO] completed")
	return nil
}

func (c *Client) getLocators(ctx context.Context, endpoint string) ([]Locator, error) {
	var out struct {
		Locators []Locator `json:"value"`
	}
	if err := c.get(ctx, endpoint, &out); err != nil {
		return nil, err
	}

	return out.Locators, nil
}

func (c *Client) GetLocators(ctx context.Context) ([]Locator, error) {
	return c.getLocators(ctx, locatorsEndpoint)
}

func (c *Client) GetLocatorsWithAsset(ctx context.Context, assetID string) ([]Locator, error) {
	endpoint := path.Join(toAssetResource(assetID), locatorsEndpoint)
	return c.getLocators(ctx, endpoint)
}

func toLocatorResource(locatorID string) string {
	return toResource(locatorsEndpoint, locatorID)
}
