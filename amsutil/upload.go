package amsutil

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/recruit-tech/go-ams"
)

const (
	uploadPolicyName       = "UploadPolicy"
	uploadDurationInMinute = 440.0
)

func UploadFile(ctx context.Context, client *ams.Client, file *os.File) (*ams.Asset, error) {
	if client == nil {
		return nil, errors.New("client missing")
	}
	if file == nil {
		return nil, errors.New("file missing")
	}

	blob, err := ams.NewFileBlob(file)
	if err != nil {
		return nil, errors.Wrap(err, "blob construct failed")
	}

	_, filename := path.Split(file.Name())
	mimeType := mime.TypeByExtension(path.Ext(filename))
	if !strings.HasPrefix(mimeType, "video/") {
		return nil, errors.Errorf("invalid file type. expected video/*, actual '%s'", mimeType)
	}

	asset, err := client.CreateAsset(ctx, filename)
	if err != nil {
		return nil, errors.Wrapf(err, "create asset failed. name='%s'", filename)
	}

	assetFile, err := client.CreateAssetFile(ctx, asset.ID, filename, mimeType)
	if err != nil {
		return nil, errors.Wrapf(err, "create asset file failed. assetID='%s'", asset.ID)
	}

	accessPolicy, err := client.CreateAccessPolicy(ctx, uploadPolicyName, uploadDurationInMinute, ams.PermissionWrite)
	if err != nil {
		return nil, errors.Wrap(err, "create access policy failed")
	}
	defer client.DeleteAccessPolicy(ctx, accessPolicy.ID)

	startTime := time.Now().Add(-5 * time.Minute)
	locator, err := client.CreateLocator(ctx, accessPolicy.ID, asset.ID, startTime, ams.LocatorSAS)
	if err != nil {
		return nil, errors.Wrap(err, "create locator failed")
	}
	defer client.DeleteLocator(ctx, locator.ID)

	uploadURL, err := locator.ToUploadURL(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "upload url build failed. name='%s'", uploadURL.String())
	}

	var blockList []string
	blockID := "block-id-01"
	if err := client.PutBlob(ctx, uploadURL, blob, blockID); err != nil {
		return nil, errors.Wrap(err, "put blob failed")
	}
	blockList = append(blockList, blockID)
	if err := client.PutBlockList(ctx, uploadURL, blockList); err != nil {
		return nil, errors.Wrap(err, "put block list failed")
	}

	assetFile.ContentFileSize = fmt.Sprint(blob.Size())
	if err := client.UpdateAssetFile(ctx, assetFile); err != nil {
		return nil, errors.Wrap(err, "update asset file failed")
	}

	return asset, nil
}
