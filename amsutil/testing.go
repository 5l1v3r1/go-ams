package amsutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/recruit-tech/go-ams"
)

func testConfigFromFile(t *testing.T, rpath string) *ams.Config {
	baseDir := os.Getenv("AMS_TEST_DIR")
	configPath := filepath.Join(baseDir, rpath)
	config, err := ams.NewConfigFromFile(configPath)
	if err != nil {
		if apath, e := filepath.Abs(configPath); e != nil {
			t.Errorf("get abspath failed %v: %v", configPath, e)
			t.Fatalf("config load failed %v: %v", configPath, err)
		} else {
			t.Fatalf("config load failed %v: %v", apath, err)
		}
	}
	config.BaseDir = baseDir
	return config
}
