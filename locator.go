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
		return nil, errors.Wrapf(err, "parse url failed: %s", l.Path)
	}
	uploadURL.Path = path.Join(uploadURL.Path, name)
	return uploadURL, nil
}

func (l *Locator) toResource() string {
	return toResource(locatorsEndpoint, l.ID)
}

func (c *Client) CreateLocator(ctx context.Context, accessPolicyID, assetID string, startTime time.Time, locatorType int) (*Locator, error) {
	params := map[string]interface{}{
		"AccessPolicyId": accessPolicyID,
		"AssetId":        assetID,
		"StartTime":      formatTime(startTime),
		"Type":           locatorType,
	}
	req, err := c.newRequest(ctx, http.MethodPost, locatorsEndpoint, withJSON(params))
	if err != nil {
		return nil, errors.Wrap(err, "request build failed")
	}

	c.logger.Printf("[INFO] create locator ...")
	var out Locator
	if err := c.do(req, http.StatusCreated, &out); err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	c.logger.Printf("[INFO] completed, new locator[#%s]", out.ID)
	return &out, nil
}

func (c *Client) DeleteLocator(ctx context.Context, locatorID string) error {
	endpoint := toLocatorResource(locatorID)
	req, err := c.newRequest(ctx, http.MethodDelete, endpoint)
	if err != nil {
		return errors.Wrap(err, "request build failed")
	}
	c.logger.Printf("[INFO] delete locator #%s ...", locatorID)
	if err := c.do(req, http.StatusNoContent, nil); err != nil {
		return errors.Wrap(err, "request failed")
	}
	c.logger.Printf("[INFO] completed")
	return nil
}

func toLocatorResource(locatorID string) string {
	return toResource(locatorsEndpoint, locatorID)
}
