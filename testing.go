package ams

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/oauth2"
)

func testTokenSource() oauth2.TokenSource {
	return oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: "<<DUMMY ACCESS TOKEN>>",
		TokenType:   "Bearer",
	})
}

func TestConfigFromFile(t *testing.T, rpath string) *Config {
	baseDir := os.Getenv("AMS_TEST_DIR")
	configPath := filepath.Join(baseDir, rpath)
	config, err := NewConfigFromFile(configPath)
	if err != nil {
		apath, _ := filepath.Abs(configPath)
		t.Fatalf("config load failed %v: %v", apath, err)
	}
	config.BaseDir = baseDir
	return config
}
