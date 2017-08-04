package amsutil

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/recruit-tech/go-ams/testutil"
)

func TestUploadFile(t *testing.T) {
	ctx := context.TODO()
	cnf, err := testutil.LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("config load failed: %v", err)
	}
	AMS, err := cnf.Client(ctx)
	if err != nil {
		t.Fatalf("client build failed")
	}

	testFile, err := os.Open(path.Join(cnf.RepoDir, "testdata", "small.mp4"))
	if err != nil {
		t.Fatalf("test file open failed: %v", err)
	}
	asset, err := UploadFile(ctx, AMS, testFile)
	if err != nil {
		t.Errorf("file uploading failed: %v", err)
	}
	if asset == nil {
		t.Errorf("return invalid asset")
	}

	if err := AMS.DeleteAsset(ctx, asset.ID); err != nil {
		t.Errorf("asset delete failed: %v", err)
	}
}
