package amsutil

import (
	"context"

	"github.com/pkg/errors"
	"github.com/recruit-tech/go-ams"
)

func Encode(ctx context.Context, client *ams.Client, assetID, mediaProcessorID, configuration string) (string, error) {
	if client == nil {
		return "", errors.New("missing client")
	}
	if len(assetID) == 0 {
		return "", errors.New("missing assetID")
	}
	if len(mediaProcessorID) == 0 {
		return "", errors.New("missing mediaProcessorID")
	}
	if len(configuration) == 0 {
		return "", errors.New("missing configuration")
	}

	asset, err := client.GetAsset(ctx, assetID)
	if err != nil {
		return "", errors.Wrapf(err, "get asset failed. assetID='%s'", assetID)
	}

	job, err := client.EncodeAsset(ctx, asset.ID, "[ENCODED]"+asset.Name, mediaProcessorID, configuration)
	if err != nil {
		return "", errors.Wrap(err, "encode asset failed")
	}

	outputMediaAssets, err := client.GetOutputMediaAssets(ctx, job.ID)
	if err != nil {
		return "", errors.Wrap(err, "get output media assets failed")
	}

	if err := client.WaitJob(ctx, job.ID); err != nil {
		return "", errors.Wrap(err, "wait job failed")
	}

	return outputMediaAssets[0].ID, nil
}
