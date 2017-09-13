package ams

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"net/url"
	"os"
	"text/template"

	"github.com/pkg/errors"
)

type Blob interface {
	FixedSizeReader
}

type FileBlob struct {
	file *os.File
	stat os.FileInfo
}

func NewFileBlob(file *os.File) (*FileBlob, error) {
	if file == nil {
		return nil, errors.New("missing file")
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "get file stat failed")
	}
	return &FileBlob{
		file: file,
		stat: stat,
	}, nil
}

func (f *FileBlob) Read(p []byte) (int, error) {
	return f.file.Read(p)
}

func (f *FileBlob) Size() int64 {
	return f.stat.Size()
}

func (f *FileBlob) Name() string {
	return f.stat.Name()
}

type BytesBlob struct {
	name string
	r    *bytes.Reader
}

func NewBytesBlob(name string, b []byte) *BytesBlob {
	return &BytesBlob{
		name: name,
		r:    bytes.NewReader(b),
	}
}

func (b *BytesBlob) Read(p []byte) (int, error) {
	return b.r.Read(p)
}

func (b *BytesBlob) Size() int64 {
	return int64(b.r.Len())
}

func (b *BytesBlob) Name() string {
	return b.name
}

func (c *Client) PutBlob(ctx context.Context, uploadURL *url.URL, blob Blob, blockID string) error {
	if uploadURL == nil {
		return errors.New("missing uploadURL")
	}
	if blob == nil {
		return errors.New("missing blob")
	}
	if len(blockID) == 0 {
		return errors.New("missing blockID")
	}

	params := url.Values{}
	params.Set("comp", "block")
	params.Set("blockid", buildBlockID(blockID))

	req, err := c.newStorageRequest(ctx, http.MethodPut, *uploadURL,
		withQuery(params),
		withCustomHeader("x-ms-blob-type", "BlockBlob"),
		withContentType("application/octet-stream"),
		withBody(blob),
	)
	if err != nil {
		return errors.Wrap(err, "request build failed")
	}
	req.ContentLength = blob.Size()

	c.logger.Printf("[INFO] put blob ...")

	if err := c.doWithClient(http.DefaultClient, req, http.StatusCreated, nil); err != nil {
		return errors.Wrap(err, "request failed")
	}
	c.logger.Printf("[INFO] completed")
	return nil
}

func (c *Client) PutBlockList(ctx context.Context, uploadURL *url.URL, blockList []string) error {
	params := url.Values{}
	params.Set("comp", "blocklist")

	blockListXML, err := buildBlockListXML(blockList)
	if err != nil {
		return errors.Wrap(err, "block list XML build failed")
	}
	req, err := c.newStorageRequest(ctx, http.MethodPut, *uploadURL, withQuery(params), withBytes(blockListXML))
	if err != nil {
		return errors.Wrap(err, "request build failed")
	}

	c.logger.Printf("[INFO] put block list ...")
	if err := c.doWithClient(http.DefaultClient, req, http.StatusCreated, nil); err != nil {
		return errors.Wrap(err, "request failed")
	}
	c.logger.Printf("[INFO] completed")

	return nil
}

var blockListXML *template.Template = template.Must(
	template.New("blockListXML").
		Funcs(template.FuncMap{"buildBlockID": buildBlockID}).
		Parse(`<?xml version="1.0" encoding="utf-8"?><BlockList>{{ range . }}<Latest>{{ . | buildBlockID }}</Latest>{{ end }}</BlockList>`),
)

func buildBlockID(blockID string) string {
	return base64.StdEncoding.EncodeToString([]byte(blockID))
}

func buildBlockListXML(blockList []string) ([]byte, error) {
	var b bytes.Buffer
	if err := blockListXML.Execute(&b, blockList); err != nil {
		return nil, errors.Wrap(err, "template execute failed")
	}
	return b.Bytes(), nil
}
