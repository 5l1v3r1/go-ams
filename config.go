package ams

import (
	"context"
	"encoding/json"
	"os"

	"github.com/orisano/go-adal"
	"github.com/pkg/errors"
)

type Config struct {
	ClientID   string
	Tenant     string
	AMSBaseURL string

	Debug bool

	ClientSecret string `json:"-"`

	BaseDir string `json:"-"`
}

func NewConfigFromFile(filepath string) (*Config, error) {
	clientSecret := os.Getenv("AAD_TOKEN")
	if len(clientSecret) == 0 {
		return nil, errors.New("missing AAD_TOKEN")
	}

	f, err := os.Open(filepath)
	if err != nil {
		return nil, errors.Wrap(err, "file open failed")
	}
	defer f.Close()

	var config Config
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return nil, errors.Wrap(err, "config file decode failed")
	}
	config.ClientSecret = clientSecret

	return &config, nil
}

func (c *Config) Client(ctx context.Context, opts ...clientOption) (*Client, error) {
	ac, err := adal.NewAuthenticationContext(c.Tenant)
	if err != nil {
		return nil, errors.Wrap(err, "authentication context construct failed")
	}
	httpClient, err := ac.Client(ctx, Resource, c.ClientID, c.ClientSecret)
	if err != nil {
		return nil, errors.Wrap(err, "httpClient construct failed")
	}

	opts = append([]clientOption{SetDebug(c.Debug)}, opts...)
	client, err := NewClient(c.AMSBaseURL, httpClient, opts...)
	if err != nil {
		return nil, err
	}
	return client, nil
}
