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
		return nil, errors.Wrapf(err, "file open failed: %s", filepath)
	}
	defer f.Close()

	var config Config
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return nil, errors.Wrap(err, "config file decode failed")
	}
	config.ClientSecret = clientSecret

	return &config, nil
}

func (c *Config) Client(ctx context.Context) (*Client, error) {
	ac, err := adal.NewAuthenticationContext(c.Tenant)
	if err != nil {
		return nil, errors.Wrap(err, "authentication context construct failed")
	}
	ts := ac.TokenSourceFromClientCredentials(ctx, Resource, c.ClientID, c.ClientSecret)
	client, err := NewClient(c.AMSBaseURL, ts)
	if err != nil {
		return nil, err
	}
	client.SetDebug(c.Debug)
	return client, err
}
