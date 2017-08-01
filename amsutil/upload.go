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

func UploadFile(ctx context.Context, client *ams.Client, file *os.File) error {
	if client == nil {
		return errors.New("client missing")
	}
	if file == nil {
		return errors.New("file missing")
	}

	stat, err := file.Stat()
	if err != nil {
		return errors.Wrapf(err, "upload file stat read failed")
	}

	_, filename := path.Split(file.Name())
	mimeType := mime.TypeByExtension(path.Ext(filename))
	if !strings.HasPrefix(mimeType, "video/") {
		return errors.Errorf("invalid file type. expected video/*, actual '%s'", mimeType)
	}

	asset, err := client.CreateAsset(ctx, filename)
	if err != nil {
		return errors.Wrapf(err, "create asset failed. name='%s'", filename)
	}

	assetFile, err := client.CreateAssetFile(ctx, asset.ID, filename, mimeType)
	if err != nil {
		return errors.Wrapf(err, "create asset file failed. assetID='%s'", asset.ID)
	}

	accessPolicy, err := client.CreateAccessPolicy(ctx, uploadPolicyName, uploadDurationInMinute, ams.PermissionWrite)
	if err != nil {
		return errors.Wrap(err, "create access policy failed")
	}
	defer client.DeleteAccessPolicy(ctx, accessPolicy.ID)

	startTime := time.Now().Add(-5 * time.Minute)
	locator, err := client.CreateLocator(ctx, accessPolicy.ID, asset.ID, startTime, ams.LocatorSAS)
	if err != nil {
		return errors.Wrap(err, "create locator failed")
	}
	defer client.DeleteLocator(ctx, locator.ID)

	uploadURL, err := locator.ToUploadURL(filename)
	if err != nil {
		return errors.Wrapf(err, "upload url build failed. name='%s'", uploadURL.String())
	}

	blockList, err := client.PutBlob(ctx, uploadURL, file)
	if err != nil {
		return errors.Wrap(err, "put blob failed")
	}

	if err := client.PutBlockList(ctx, uploadURL, blockList); err != nil {
		return errors.Wrap(err, "put block list failed")
	}

	assetFile.ContentFileSize = fmt.Sprint(stat.Size())
	if err := client.UpdateAssetFile(ctx, assetFile); err != nil {
		return errors.Wrap(err, "update asset file failed")
	}

	return nil
}
