package ams

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func setStorageDefaultHeader(req *http.Request) {
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Date", time.Now().UTC().Format(time.RFC3339))
	req.Header.Set("x-ms-version", "2017-04-17")
}

func (c *Client) PutBlob(uploadURL *url.URL, file *os.File) ([]int, error) {
	return c.PutBlobWithContext(context.Background(), uploadURL, file)
}

func (c *Client) PutBlobWithContext(ctx context.Context, uploadURL *url.URL, file *os.File) ([]int, error) {
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "uploading file stat read failed")
	}

	u := *uploadURL
	query := u.Query()
	query.Add("comp", "block")
	query.Add("blockid", buildBlockID(1))
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodPut, u.String(), file)
	if err != nil {
		return nil, errors.Wrap(err, "put blob request build failed")
	}
	setStorageDefaultHeader(req)
	req.Header.Set("x-ms-blob-type", "BlockBlob")
	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = fileInfo.Size()

	if err := c.do(req, http.StatusCreated, nil); err != nil {
		return nil, errors.Wrap(err, "put blob request failed")
	}
	return []int{1}, nil
}

func (c *Client) PutBlockList(uploadURL *url.URL, blockList []int) error {
	return c.PutBlockListWithContext(context.Background(), uploadURL, blockList)
}

func (c *Client) PutBlockListWithContext(ctx context.Context, uploadURL *url.URL, blockList []int) error {
	u := *uploadURL
	query := u.Query()
	query.Add("comp", "blocklist")
	u.RawQuery = query.Encode()

	blockListXML := buildBlockListXML(blockList)
	body := strings.NewReader(blockListXML)

	req, err := http.NewRequest(http.MethodPut, u.String(), body)
	if err != nil {
		return errors.Wrap(err, "put block list request build failed")
	}
	setStorageDefaultHeader(req)
	req.ContentLength = int64(len(blockListXML))

	if err := c.do(req, http.StatusCreated, nil); err != nil {
		return errors.Wrap(err, "put block list request failed")
	}

}

func buildBlockID(blockID int) string {
	s := fmt.Sprintf("BlockId%07d", blockID)
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func buildBlockListXML(blockList []int) string {
	header := "<?xml version=\"1.0\" encoding=\"utf-8\"?><BlockList>"
	footer := "</BlockList>"
	body := ""
	for _, blockID := range blockList {
		body += fmt.Sprintf("<Latest>%s</Latest>", buildBlockID(blockID))
	}
	return header + body + footer
}
