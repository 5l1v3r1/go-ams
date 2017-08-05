package testutil

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/orisano/go-adal"
	"github.com/pkg/errors"
	"github.com/recruit-tech/go-ams"
)

type Config struct {
	ClientID     string
	ClientSecret string `json:"-"`

	Tenant     string
	AMSBaseURL string

	RepoDir string `json:"-"`
}

func LoadConfigFromEnv() (*Config, error) {
	repoDir := os.Getenv("AMS_REPO_DIR")
	if len(repoDir) == 0 {
		return nil, errors.New("missing AMS_REPO_DIR")
	}
	clientSecret := os.Getenv("AAD_TOKEN")
	if len(clientSecret) == 0 {
		return nil, errors.New("missing AAD_TOKEN")
	}

	f, err := os.Open(filepath.Join(repoDir, "test_config.json"))
	if err != nil {
		return nil, errors.Wrap(err, "test_config.json open failed")
	}
	defer f.Close()

	var config Config
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return nil, errors.Wrap(err, "config file decode failed")
	}

	config.ClientSecret = clientSecret
	config.RepoDir = repoDir
	return &config, nil
}

func (c *Config) Client(ctx context.Context) (*ams.Client, error) {
	ac, err := adal.NewAuthenticationContext(c.Tenant)
	if err != nil {
		return nil, errors.Wrap(err, "authentication context construct failed")
	}
	ts := ac.TokenSourceFromClientCredentials(ctx, ams.Resource, c.ClientID, c.ClientSecret)
	return ams.NewClient(c.AMSBaseURL, ts)
}
