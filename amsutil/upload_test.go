package amsutil

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestUploadFile(t *testing.T) {
	ctx := context.TODO()
	cnf := testConfigFromFile(t, "config.json")
	AMS, err := cnf.Client(ctx)
	if err != nil {
		t.Fatalf("client build failed")
	}
	testFile, err := os.Open(filepath.Join(cnf.BaseDir, "testdata", "small.mp4"))
	if err != nil {
		t.Fatalf("video file open failed: %v", err)
	}
	defer testFile.Close()

	asset, err := UploadFile(ctx, AMS, testFile, 4*1024*1024, 5)
	if err != nil {
		t.Errorf("file uploading failed: %v", err)
	}
	if asset == nil {
		t.Fatal("return invalid asset")
	}

	if err := AMS.DeleteAsset(ctx, asset.ID); err != nil {
		t.Errorf("asset delete failed: %v", err)
	}
}
