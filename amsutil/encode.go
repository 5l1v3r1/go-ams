package amsutil

import (
	"context"

	"github.com/pkg/errors"
	"github.com/recruit-tech/go-ams"
)

func Encode(ctx context.Context, client *ams.Client, assetID, mediaProcessorID, configuration string) ([]ams.Asset, *ams.Job, error) {
	if client == nil {
		return nil, nil, errors.New("missing client")
	}
	if len(assetID) == 0 {
		return nil, nil, errors.New("missing assetID")
	}
	if len(mediaProcessorID) == 0 {
		return nil, nil, errors.New("missing mediaProcessorID")
	}
	if len(configuration) == 0 {
		return nil, nil, errors.New("missing configuration")
	}

	asset, err := client.GetAsset(ctx, assetID)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "get asset failed. assetID='%s'", assetID)
	}

	job, err := client.AddEncodeJob(ctx, asset.ID, "[ENCODED]"+asset.Name, mediaProcessorID, configuration)
	if err != nil {
		return nil, nil, errors.Wrap(err, "encode asset failed")
	}

	outputMediaAssets, err := client.GetOutputMediaAssets(ctx, job.ID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "get output media assets failed")
	}

	return outputMediaAssets, job, nil
}
