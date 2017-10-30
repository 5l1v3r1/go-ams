package amsutil

import (
	"context"

	"github.com/pkg/errors"
	"github.com/recruit-tech/go-ams"
)

func Encode(ctx context.Context, client *ams.Client, assetID, mediaProcessorID, configuration string) ([]ams.Asset, *ams.Job, error) {
	if ctx == nil {
		return nil, nil, errors.New("missing ctx")
	}
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
		return nil, nil, errors.Wrapf(err, "failed to get asset. assetID='%v'", assetID)
	}

	job, err := client.AddEncodeJob(ctx, asset.ID, mediaProcessorID, "")
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to encode asset. assetID='%v'", asset.ID)
	}

	outputMediaAssets, err := client.GetOutputMediaAssets(ctx, job.ID)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to get output media assets. jobID='%v'", job.ID)
	}

	return outputMediaAssets, job, nil
}
