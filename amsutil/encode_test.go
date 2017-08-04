package amsutil

import (
	"context"
	"testing"

	"os"
	"path"

	"github.com/recruit-tech/go-ams/testutil"
)

func TestEncode(t *testing.T) {
	ctx := context.TODO()
	cnf, err := testutil.LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("config load failed: %v", err)
	}
	AMS, err := cnf.Client(ctx)
	if err != nil {
		t.Fatalf("client construct failed: %v", err)
	}

	f, err := os.Open(path.Join(cnf.RepoDir, "testdata", "small.mp4"))
	if err != nil {
		t.Fatalf("video file open failed: %v", err)
	}
	defer f.Close()

	asset, err := UploadFile(ctx, AMS, f)
	if err != nil {
		t.Fatalf("file uploading failed: %v", err)
	}

	mediaProcessors, err := AMS.GetMediaProcessors(ctx)
	if err != nil {
		t.Fatalf("get media processors failed: %v", err)
	}

	var MES string
	for _, mediaProcessor := range mediaProcessors {
		if mediaProcessor.Name == "Media Encoder Standard" {
			MES = mediaProcessor.ID
			break
		}
	}

	resultAssetID, err := Encode(ctx, AMS, asset.ID, MES, "Adaptive Streaming")
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	if err := AMS.DeleteAsset(ctx, asset.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	if err := AMS.DeleteAsset(ctx, resultAssetID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
}
