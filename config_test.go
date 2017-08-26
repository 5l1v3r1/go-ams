package ams

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
)

func TestNewConfigFromFile(t *testing.T) {
	t.Run("envNotSet", func(t *testing.T) {
		tf, tfClose := testTempFile(t)
		defer tfClose()

		if err := ioutil.WriteFile(tf, []byte(`{"test":"a"}`), 0644); err != nil {
			t.Fatal(err)
		}

		orig := os.Getenv("AAD_TOKEN")
		defer os.Setenv("AAD_TOKEN", orig)

		os.Setenv("AAD_TOKEN", "")

		conf, err := NewConfigFromFile(tf)
		if err == nil {
			t.Error("accept invalid AAD_TOKEN")
		}
		if conf != nil {
			t.Error("return invalid config")
		}
	})
	t.Run("missingFilepath", func(t *testing.T) {
		orig := os.Getenv("AAD_TOKEN")
		defer os.Setenv("AAD_TOKEN", orig)

		os.Setenv("AAD_TOKEN", "sample aad token")

		conf, err := NewConfigFromFile("")
		if err == nil {
			t.Error("accept invalid filepath")
		}
		if conf != nil {
			t.Error("return invalid config")
		}
	})
	t.Run("brokenConfig", func(t *testing.T) {
		tf, tfClose := testTempFile(t)
		defer tfClose()

		rawConfig := `{"AMSBaseURL":"https://fake.url/api", "ClientID":"broken-id", "Tenant":"broken.tenant.com",}`
		if err := ioutil.WriteFile(tf, []byte(rawConfig), 0644); err != nil {
			t.Fatal(err)
		}

		orig := os.Getenv("AAD_TOKEN")
		defer os.Setenv("AAD_TOKEN", orig)
		os.Setenv("AAD_TOKEN", "sample aad token")

		conf, err := NewConfigFromFile(tf)
		if err == nil {
			t.Error("accept invalid config file")
		}
		if conf != nil {
			t.Error("return invalid config")
		}
	})
	t.Run("positiveCase", func(t *testing.T) {
		tf, tfClose := testTempFile(t)
		defer tfClose()

		rawConfig := `{"AMSBaseURL":"https://fake.url/api", "ClientID":"sample-id", "Tenant":"sample.tenant.com"}`
		if err := ioutil.WriteFile(tf, []byte(rawConfig), 0644); err != nil {
			t.Fatal(err)
		}

		clientSecret := "sample-aad-token"

		orig := os.Getenv("AAD_TOKEN")
		defer os.Setenv("AAD_TOKEN", orig)
		os.Setenv("AAD_TOKEN", clientSecret)

		conf, err := NewConfigFromFile(tf)
		if err != nil {
			t.Error(err)
		}
		if conf == nil {
			t.Error("return nil config")
		}
		baseURL := "https://fake.url/api"
		if conf.AMSBaseURL != baseURL {
			t.Errorf("unexpected AMSBaseURL. expected: %v, actual: %v", baseURL, conf.AMSBaseURL)
		}

		clientID := "sample-id"
		if conf.ClientID != clientID {
			t.Errorf("unexpected ClientID. expected: %v, actual: %v", clientID, conf.ClientID)
		}

		tenant := "sample.tenant.com"
		if conf.Tenant != tenant {
			t.Errorf("unexpected Tenant. expected: %v, actual: %v", tenant, conf.Tenant)
		}

		if conf.ClientSecret != clientSecret {
			t.Errorf("unexpected ClientSecret. expected: %v, actual: %v", clientSecret, conf.ClientSecret)
		}
	})
}

func TestConfig_Client(t *testing.T) {

	t.Run("invalidTenant", func(t *testing.T) {
		config := Config{
			ClientID:     "dummy-client-id",
			Tenant:       "",
			AMSBaseURL:   "http://dummy.url/api/",
			ClientSecret: "dummy-client-secret",
		}
		client, err := config.Client(context.TODO())
		if err == nil {
			t.Error("accept invalid tenant")
		}
		if client != nil {
			t.Error("return invalid client")
		}
	})
	t.Run("invalidAMSBaseURL", func(t *testing.T) {
		config := Config{
			ClientID:     "dummy-client-id",
			Tenant:       "dummy.tenant.com",
			AMSBaseURL:   "dummy.url",
			ClientSecret: "dummy-client-secret",
		}
		client, err := config.Client(context.TODO())
		if err == nil {
			t.Error("accept invalid base url")
		}
		if client != nil {
			t.Error("return invalid client")
		}
	})
	t.Run("positiveCase", func(t *testing.T) {
		config := Config{
			ClientID:     "dummy-client-id",
			Tenant:       "dummy.tenant.com",
			AMSBaseURL:   "http://dummy.url/api",
			ClientSecret: "dummy-client-secret",
		}
		client, err := config.Client(context.TODO())
		if err != nil {
			t.Fatal(err)
		}
		if client == nil {
			t.Error("return nil client")
		}
	})
}
