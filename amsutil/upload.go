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

type Uploadable interface {
	Name() string
	Size() int64
	Blobs() []ams.Blob
}

type uploadable struct {
	name  string
	size  int64
	blobs []ams.Blob
}

func (u *uploadable) Name() string      { return u.name }
func (u *uploadable) Size() int64       { return u.size }
func (u *uploadable) Blobs() []ams.Blob { return u.blobs }

func newUploadbleSingle(name string, blob ams.Blob) *uploadable {
	return &uploadable{
		name:  name,
		size:  blob.Size(),
		blobs: []ams.Blob{blob},
	}
}

func UploadFile(ctx context.Context, client *ams.Client, file *os.File) (*ams.Asset, error) {
	if client == nil {
		return nil, errors.New("client missing")
	}
	if file == nil {
		return nil, errors.New("file missing")
	}

	blob, err := ams.NewFileBlob(file)
	if err != nil {
		return nil, errors.Wrap(err, "file blob construct failed")
	}

	mimeType := mime.TypeByExtension(path.Ext(blob.Name()))
	if !strings.HasPrefix(mimeType, "video/") {
		return nil, errors.Errorf("invalid file type. expected video/*, actual '%v'", mimeType)
	}

	return Upload(ctx, client, newUploadbleSingle(blob.Name(), blob), mimeType)
}

func Upload(ctx context.Context, client *ams.Client, uploadable Uploadable, mimeType string) (*ams.Asset, error) {
	if client == nil {
		return nil, errors.New("client missing")
	}
	if uploadable == nil {
		return nil, errors.New("uploadable missing")
	}

	name := uploadable.Name()
	asset, err := client.CreateAsset(ctx, name)
	if err != nil {
		return nil, errors.Wrapf(err, "create asset failed. name='%s'", name)
	}

	assetFile, err := client.CreateAssetFile(ctx, asset.ID, name, mimeType)
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

	uploadURL, err := locator.ToUploadURL(name)
	if err != nil {
		return nil, errors.Wrapf(err, "upload url build failed")
	}

	var blockList []string
	for i, blob := range uploadable.Blobs() {
		blockID := fmt.Sprintf("block-id-%v", i+1)
		if err := client.PutBlob(ctx, uploadURL, blob, blockID); err != nil {
			return nil, errors.Wrap(err, "put blob failed")
		}
		blockList = append(blockList, blockID)
	}

	if err := client.PutBlockList(ctx, uploadURL, blockList); err != nil {
		return nil, errors.Wrap(err, "put block list failed")
	}

	assetFile.ContentFileSize = fmt.Sprint(uploadable.Size())
	if err := client.UpdateAssetFile(ctx, assetFile); err != nil {
		return nil, errors.Wrap(err, "update asset file failed")
	}

	return asset, nil
}
