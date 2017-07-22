package ams

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type AccessPolicy struct {
	ID                string  `json:"Id"`
	Created           string  `json:"Created"`
	LastModified      string  `json:"LastModified"`
	Name              string  `json:"Name"`
	DurationInMinutes float64 `json:"DurationInMinutes"`
	Permissions       int     `json:"Permissions"`
}

func (c *Client) CreateAccessPolicy(name, durationInMinutes, permissions string) (*AccessPolicy, error) {
	return c.CreateAccessPolicyWithContext(context.Background(), name, durationInMinutes, permissions)
}

func (c *Client) CreateAccessPolicyWithContext(ctx context.Context, name, durationInMinutes, permissions string) (*AccessPolicy, error) {
	params := map[string]interface{}{
		"Name":              name,
		"DurationInMinutes": durationInMinutes,
		"Permissions":       permissions,
	}
	body, err := encodeParams(params)
	if err != nil {
		return nil, errors.Wrap(err, "create access-policy parameter encode failed")
	}
	req, err := c.newRequest(ctx, http.MethodPost, "AccessPolicies", body)
	if err != nil {
		return nil, errors.Wrap(err, "create access-policy request build failed")
	}
	var out AccessPolicy
	if err := c.do(req, http.StatusCreated, &out); err != nil {
		return nil, errors.Wrap(err, "create access-policy request failed")
	}
	return &out, nil
}

func (c *Client) DeleteAccessPolicy(accessPolicy *AccessPolicy) error {
	return c.DeleteAccessPolicyWithContext(context.Background(), accessPolicy)
}

func (c *Client) DeleteAccessPolicyWithContext(ctx context.Context, accessPolicy *AccessPolicy) error {
	endpoint := fmt.Sprintf("AccessPolicies('%s')", accessPolicy.ID)
	req, err := c.newRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return errors.Wrap(err, "delete access-policy request build failed")
	}
	if err := c.do(req, http.StatusNoContent, nil); err != nil {
		return errors.Wrap(err, "delete access-policy request failed")
	}
	return nil
}
