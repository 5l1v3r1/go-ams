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

func (a *AccessPolicy) toResource() string {
	return toResource(accessPoliciesEndpoint, a.ID)
}

func (c *Client) CreateAccessPolicyWithContext(ctx context.Context, name string, durationInMinutes float64, permissions int) (*AccessPolicy, error) {
	params := map[string]interface{}{
		"Name":              name,
		"DurationInMinutes": durationInMinutes,
		"Permissions":       permissions,
	}
	body, err := encodeParams(params)
	if err != nil {
		return nil, errors.Wrap(err, "parameter encode failed")
	}
	req, err := c.newRequest(ctx, http.MethodPost, accessPoliciesEndpoint, body)
	if err != nil {
		return nil, errors.Wrap(err, "request build failed")
	}
	var out AccessPolicy
	if err := c.do(req, http.StatusCreated, &out); err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	return &out, nil
}

func (c *Client) DeleteAccessPolicyWithContext(ctx context.Context, accessPolicy *AccessPolicy) error {
	req, err := c.newRequest(ctx, http.MethodDelete, accessPolicy.toResource(), nil)
	if err != nil {
		return errors.Wrap(err, "request build failed")
	}
	if err := c.do(req, http.StatusNoContent, nil); err != nil {
		return errors.Wrap(err, "request failed")
	}
	return nil
}
