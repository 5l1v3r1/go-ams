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

type AccessPolicy struct {
	ID                string  `json:"Id"`
	Created           string  `json:"Created"`
	LastModified      string  `json:"LastModified"`
	Name              string  `json:"Name"`
	DurationInMinutes float64 `json:"DurationInMinutes"`
	Permissions       int     `json:"Permissions"`
}

func (c *Client) CreateAccessPolicy(ctx context.Context, name string, durationInMinutes float64, permissions int) (*AccessPolicy, error) {
	c.logger.Printf("[INFO] create access policy [name=%#v,permissions=%d] ...", name, permissions)

	params := map[string]interface{}{
		"Name":              name,
		"DurationInMinutes": durationInMinutes,
		"Permissions":       permissions,
	}
	var out AccessPolicy
	if err := c.post(ctx, accessPoliciesEndpoint, params, &out); err != nil {
		return nil, err
	}

	c.logger.Printf("[INFO] completed, new access policy[#%s]", out.ID)
	return &out, nil
}

func (c *Client) DeleteAccessPolicy(ctx context.Context, accessPolicyID string) error {
	endpoint := toAccessPolicyResource(accessPolicyID)
	req, err := c.newRequest(ctx, http.MethodDelete, endpoint)
	if err != nil {
		return errors.Wrap(err, "request build failed")
	}
	c.logger.Printf("[INFO] delete access policy #%s ...", accessPolicyID)
	if err := c.do(req, http.StatusNoContent, nil); err != nil {
		return errors.Wrap(err, "request failed")
	}
	c.logger.Printf("[INFO] completed")
	return nil
}

func toAccessPolicyResource(accessPolicyID string) string {
	return toResource(accessPoliciesEndpoint, accessPolicyID)
}
