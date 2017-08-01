package ams

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"text/template"

	"github.com/pkg/errors"
)

func (c *Client) PutBlob(ctx context.Context, uploadURL *url.URL, file *os.File) ([]int, error) {
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "uploading file stat read failed")
	}
	params := url.Values{
		"comp":    {"block"},
		"blockid": {buildBlockID(1)},
	}
	req, err := c.newRequest(ctx, http.MethodPut, "",
		setURL(uploadURL.String()),
		useStorageAPI(),
		withQuery(params),
		withBlobType("BlockBlob"),
		withContentType("application/octet-stream"),
		withBody(file),
	)
	if err != nil {
		return nil, errors.Wrap(err, "request build failed")
	}
	req.ContentLength = fileInfo.Size()
	if err := c.do(req, http.StatusCreated, nil); err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	return []int{1}, nil
}

func (c *Client) PutBlockList(ctx context.Context, uploadURL *url.URL, blockList []int) error {
	params := url.Values{
		"comp": {"blocklist"},
	}
	blockListXML, err := BuildBlockListXML(blockList)
	if err != nil {
		return errors.Wrap(err, "block list XML build failed")
	}
	req, err := c.newRequest(ctx, http.MethodPut, "",
		setURL(uploadURL.String()),
		useStorageAPI(),
		withQuery(params),
		withBytes(blockListXML),
	)
	if err != nil {
		return errors.Wrap(err, "request build failed")
	}
	req.ContentLength = int64(len(blockListXML))

	if err := c.do(req, http.StatusCreated, nil); err != nil {
		return errors.Wrap(err, "request failed")
	}

	return nil
}

var blockListXML *template.Template = template.Must(
	template.New("blockListXML").Funcs(template.FuncMap{
		"buildBlockID": buildBlockID,
	}).Parse(`<?xml version="1.0" encoding="utf-8"?><BlockList>{{ range . }}<Latest>{{ . | buildBlockID }}</Latest>{{ end }}</BlockList>`))

func buildBlockID(blockID int) string {
	s := fmt.Sprintf("BlockId%07d", blockID)
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func BuildBlockListXML(blockList []int) ([]byte, error) {
	var b bytes.Buffer
	if err := blockListXML.Execute(&b, blockList); err != nil {
		return nil, errors.Wrap(err, "template execute failed")
	}
	return b.Bytes(), nil
}
