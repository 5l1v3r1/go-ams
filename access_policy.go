package ams

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
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
		return nil, err
	}
	req, err := c.newRequest(ctx, http.MethodPost, "AccessPolicies", body)
	if err != nil {
		return nil, err
	}
	var out AccessPolicy
	if err := c.do(req, http.StatusCreated, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteAccessPolicy(accessPolicy *AccessPolicy) error {
	return c.DeleteAccessPolicyWithContext(context.Background(), accessPolicy)
}

func (c *Client) DeleteAccessPolicyWithContext(ctx context.Context, accessPolicy *AccessPolicy) error {
	endpoint := fmt.Sprintf("AccessPolicies('%s')", url.PathEscape(accessPolicy.ID))
	req, err := c.newRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	return c.do(req, http.StatusNoContent, nil)
}
