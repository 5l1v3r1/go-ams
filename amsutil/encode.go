package amsutil

import (
	"github.com/pkg/errors"
	"github.com/recruit-tech/go-ams"
)

func Encode(client *ams.Client, assetID, mediaProcessorID, configuration string) (string, error) {
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

	asset, err := client.GetAsset(assetID)
	if err != nil {
		return "", errors.Wrapf(err, "get asset failed. assetID='%s'", assetID)
	}

	job, err := client.EncodeAsset(mediaProcessorID, configuration, asset)
	if err != nil {
		return "", errors.Wrap(err, "encode asset failed")
	}

	outputMediaAssets, err := client.GetOutputMediaAssets(job)
	if err != nil {
		return "", errors.Wrap(err, "get output media assets failed")
	}

	if err := client.WaitJob(job); err != nil {
		return "", errors.Wrap(err, "wait job failed")
	}

	return outputMediaAssets[0].ID, nil
}
