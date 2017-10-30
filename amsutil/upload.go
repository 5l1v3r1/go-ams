package amsutil

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/recruit-tech/go-ams"
)

const (
	uploadPolicyName       = "UploadPolicy"
	uploadDurationInMinute = 440.0
)

var TimeNow func() time.Time = time.Now

type Uploadable interface {
	io.Reader
	Name() string
}

type uploadable struct {
	r    io.Reader
	name string
}

func (u *uploadable) Read(p []byte) (n int, err error) {
	return u.r.Read(p)
}

func (u *uploadable) Name() string {
	return u.name
}

func NewUploadableFile(file *os.File) (Uploadable, error) {
	if file == nil {
		return nil, errors.New("missing file")
	}
	return &uploadable{
		r:    file,
		name: filepath.Base(file.Name()),
	}, nil
}

func UploadFile(ctx context.Context, client *ams.Client, file *os.File, chunkSize int64, workers uint) (*ams.Asset, error) {
	if ctx == nil {
		return nil, errors.New("missing ctx")
	}
	if client == nil {
		return nil, errors.New("missing client")
	}
	if file == nil {
		return nil, errors.New("missing file")
	}
	if chunkSize <= 0 {
		return nil, errors.New("chunkSize must be greater than 0")
	}
	if workers == 0 {
		return nil, errors.New("workers must be greater than 0")
	}

	u, _ := NewUploadableFile(file)
	mimeType := mime.TypeByExtension(filepath.Ext(u.Name()))
	if !strings.HasPrefix(mimeType, "video/") {
		return nil, errors.Errorf("invalid file type. expected video/*, but got '%v'", mimeType)
	}

	return Upload(ctx, client, u, mimeType, chunkSize, workers)
}

func Upload(ctx context.Context, client *ams.Client, uploadable Uploadable, mimeType string, chunkSize int64, workers uint) (*ams.Asset, error) {
	if ctx == nil {
		return nil, errors.New("missing ctx")
	}
	if client == nil {
		return nil, errors.New("missing client")
	}
	if uploadable == nil {
		return nil, errors.New("missing uploadable")
	}
	if !strings.HasPrefix(mimeType, "video/") {
		return nil, errors.Errorf("invalid mime type. expected video/*, but got '%v'", mimeType)
	}
	if chunkSize <= 0 {
		return nil, errors.New("chunkSize must be greater than 0")
	}
	if workers == 0 {
		return nil, errors.New("workers must be greater than 0")
	}

	name := uploadable.Name()
	asset, err := client.CreateAsset(ctx, name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create asset. name='%s'", name)
	}

	assetFile, err := client.CreateAssetFile(ctx, asset.ID, name, mimeType)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create asset file. assetID='%s'", asset.ID)
	}

	accessPolicy, err := client.CreateAccessPolicy(ctx, uploadPolicyName, uploadDurationInMinute, ams.PermissionWrite)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create access policy")
	}
	defer client.DeleteAccessPolicy(ctx, accessPolicy.ID)

	// ref: https://docs.microsoft.com/en-US/azure/media-services/media-services-rest-upload-files
	// for clock skew
	startTime := TimeNow().Add(-5 * time.Minute)
	locator, err := client.CreateLocator(ctx, accessPolicy.ID, asset.ID, startTime, ams.LocatorSAS)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create locator")
	}
	defer client.DeleteLocator(ctx, locator.ID)

	uploadURL, err := locator.ToUploadURL(name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to construct upload url")
	}

	sasc, err := client.NewSASClient(uploadURL.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to construct SASClient")
	}

	contentLength, err := sasc.Upload(ctx, uploadable, chunkSize, workers)
	if err != nil {
		return nil, err
	}

	assetFile.ContentFileSize = fmt.Sprint(contentLength)
	if err := client.UpdateAssetFile(ctx, assetFile); err != nil {
		return nil, errors.Wrap(err, "failed to update asset file")
	}

	return asset, nil
}
