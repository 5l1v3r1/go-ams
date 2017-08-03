package ams

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
)

const (
	accessPoliciesEndpoint = "AccessPolicies"
)

const (
	PermissionRead = 1 << iota
	PermissionWrite
	PermissionDelete
	PermissionList
	PermissionNone = 0
)

const (
	LocatorNone = iota
	LocatorSAS
	LocatorOnDemandOrigin
)

type AccessPolicy struct {
	ID                string  `json:"Id"`
	Created           string  `json:"Created"`
	LastModified      string  `json:"LastModified"`
	Name              string  `json:"Name"`
	DurationInMinutes float64 `json:"DurationInMinutes"`
	Permissions       int     `json:"Permissions"`
}

func (c *Client) CreateAccessPolicy(ctx context.Context, name string, durationInMinutes float64, permissions int) (*AccessPolicy, error) {
	params := map[string]interface{}{
		"Name":              name,
		"DurationInMinutes": durationInMinutes,
		"Permissions":       permissions,
	}
	req, err := c.newRequest(ctx, http.MethodPost, accessPoliciesEndpoint, withJSON(params))
	if err != nil {
		return nil, errors.Wrap(err, "request build failed")
	}
	var out AccessPolicy
	if err := c.do(req, http.StatusCreated, &out); err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	return &out, nil
}

func (c *Client) DeleteAccessPolicy(ctx context.Context, accessPolicyID string) error {
	endpoint := toAccessPolicyResource(accessPolicyID)
	req, err := c.newRequest(ctx, http.MethodDelete, endpoint)
	if err != nil {
		return errors.Wrap(err, "request build failed")
	}
	if err := c.do(req, http.StatusNoContent, nil); err != nil {
		return errors.Wrap(err, "request failed")
	}
	return nil
}

func toAccessPolicyResource(accessPolicyID string) string {
	return toResource(accessPoliciesEndpoint, accessPolicyID)
}
