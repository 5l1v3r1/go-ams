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
)

func Publish(ctx context.Context, client *ams.Client, assetID string, minutes float64) (string, error) {
	if ctx == nil {
		return "", errors.New("missing ctx")
	}
	if client == nil {
		return "", errors.New("missing client")
	}
	if len(assetID) == 0 {
		return "", errors.New("missing assetID")
	}
	if minutes <= 0 {
		return "", errors.New("minutes must be greater than 0")
	}

	asset, err := client.GetAsset(ctx, assetID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get asset. assetID='%v'", assetID)
	}

	success := false

	accessPolicy, err := client.CreateAccessPolicy(ctx, publishAccessPolicyName, minutes, ams.PermissionRead)
	if err != nil {
		return "", errors.Wrap(err, "failed to create access policy")
	}
	defer func() {
		if !success {
			client.DeleteAccessPolicy(ctx, accessPolicy.ID)
		}
	}()

	startTime := time.Now().Add(-5 * time.Minute)
	locator, err := client.CreateLocator(ctx, accessPolicy.ID, asset.ID, startTime, ams.LocatorOnDemandOrigin)
	if err != nil {
		return "", errors.Wrap(err, "failed to create locator")
	}
	defer func() {
		if !success {
			client.DeleteLocator(ctx, locator.ID)
		}
	}()

	assetFiles, err := client.GetAssetFiles(ctx, asset.ID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get asset files")
	}

	if len(assetFiles) == 0 {
		return "", errors.Errorf("asset files not found. asset[#%v] is empty", asset.ID)
	}

	manifest := findAssetManifest(assetFiles)

	u, err := url.ParseRequestURI(locator.Path)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse locator path")
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
