package amsutil

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
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

type multiError struct {
	errs []error
}

func (m *multiError) Error() string {
	messages := make([]string, 0, len(m.errs))
	for _, err := range m.errs {
		messages = append(messages, err.Error())
	}
	return fmt.Sprintf("multiple error: %s", strings.Join(messages, ","))
}

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

func SplitBlob(blob ams.Blob, chunkSize int64) ([]ams.Blob, error) {
	size := blob.Size()
	body, err := ioutil.ReadAll(blob)
	if err != nil {
		return nil, errors.Wrap(err, "blob failed to read")
	}
	var chunk []byte
	chunks := make([]ams.Blob, 0, (size+chunkSize-1)/chunkSize)
	for int64(len(body)) >= chunkSize {
		chunk, body = body[:chunkSize], body[chunkSize:]
		chunks = append(chunks, bytes.NewReader(chunk))
	}
	if len(body) > 0 {
		chunks = append(chunks, bytes.NewReader(body))
	}
	return chunks, nil
}

func NewUploadable(name string, blob ams.Blob, chunkSize int64) (Uploadable, error) {
	blobs, err := SplitBlob(blob, chunkSize)
	if err != nil {
		return nil, errors.Wrap(err, "blob failed to split")
	}
	return &uploadable{
		name:  name,
		size:  blob.Size(),
		blobs: blobs,
	}, nil
}

func UploadFile(ctx context.Context, client *ams.Client, file *os.File, chunkSize, worker int64) (*ams.Asset, error) {
	if client == nil {
		return nil, errors.New("client missing")
	}
	if file == nil {
		return nil, errors.New("file missing")
	}

	fblob, err := ams.NewFileBlob(file)
	if err != nil {
		return nil, errors.Wrap(err, "file blob construct failed")
	}

	mimeType := mime.TypeByExtension(path.Ext(fblob.Name()))
	if !strings.HasPrefix(mimeType, "video/") {
		return nil, errors.Errorf("invalid file type. expected video/*, actual '%v'", mimeType)
	}

	uploadable, err := NewUploadable(fblob.Name(), fblob, chunkSize)
	if err != nil {
		return nil, errors.Wrap(err, "uploadable failed to construct")
	}

	return Upload(ctx, client, uploadable, mimeType, worker)
}

func Upload(ctx context.Context, client *ams.Client, uploadable Uploadable, mimeType string, worker int64) (*ams.Asset, error) {
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

	blobs := uploadable.Blobs()
	blobsSize := len(blobs)
	jobs := make(chan int, worker)
	errs := make(chan error, worker)

	for w := int64(0); w < worker; w++ {
		go func(c context.Context, id int64, in <-chan int, e chan<- error) {
			for i := range in {
				blockID := fmt.Sprintf("block-id-%v", i+1)
				e <- client.PutBlob(c, uploadURL, blobs[i], blockID)
			}
		}(ctx, w, jobs, errs)
	}

	var blockList []string
	for i := 0; i < blobsSize; i++ {
		blockID := fmt.Sprintf("block-id-%v", i+1)
		blockList = append(blockList, blockID)
		jobs <- i
	}
	close(jobs)

	var es []error
	for i := 0; i < blobsSize; i++ {
		err := <-errs
		if err != nil {
			es = append(es, err)
		}
	}

	if len(es) != 0 {
		return nil, &multiError{es}
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
