package amsutil

import (
	"context"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/recruit-tech/go-ams"
)

const (
	publishAccessPolicyName = "ViewPolicy"
	publishDurationInMinute = 43200
)

func Publish(ctx context.Context, client *ams.Client, assetID string) (string, error) {
	if client == nil {
		return "", errors.New("missing client")
	}
	if len(assetID) == 0 {
		return "", errors.New("missing assetID")
	}

	asset, err := client.GetAsset(ctx, assetID)
	if err != nil {
		return "", errors.Wrapf(err, "get asset failed. assetID='%s'", assetID)
	}

	success := false

	accessPolicy, err := client.CreateAccessPolicy(ctx, publishAccessPolicyName, publishDurationInMinute, ams.PermissionRead)
	if err != nil {
		return "", errors.Wrap(err, "create access policy failed")
	}
	defer func() {
		if !success {
			client.DeleteAccessPolicy(ctx, accessPolicy.ID)
		}
	}()

	startTime := time.Now().Add(-5 * time.Minute)
	locator, err := client.CreateLocator(ctx, accessPolicy.ID, asset.ID, startTime, ams.LocatorOnDemandOrigin)
	if err != nil {
		return "", errors.Wrap(err, "create locator failed")
	}
	defer func() {
		if !success {
			client.DeleteLocator(ctx, locator.ID)
		}
	}()

	assetFiles, err := client.GetAssetFiles(ctx, asset.ID)
	if err != nil {
		return "", errors.Wrap(err, "get asset files failed")
	}

	manifest := findAssetManifest(assetFiles)

	u, err := url.ParseRequestURI(locator.Path)
	if err != nil {
		return "", errors.Wrapf(err, "locator path parse failed. path='%s'", locator.Path)
	}

	if manifest != nil {
		u.Path = path.Join(u.Path, manifest.Name, "manifest")
	} else {
		u.Path = path.Join(u.Path, assetFiles[0].Name)
	}
	success = true
	return u.String(), nil
}

func findAssetManifest(assetFiles []ams.AssetFile) *ams.AssetFile {
	for _, assetFile := range assetFiles {
		if strings.HasSuffix(assetFile.Name, ".ism") {
			return &assetFile
		}
	}
	return nil
}
