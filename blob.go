package ams

import (
	"context"
	"io"
	"net/url"

	"github.com/pkg/errors"
	ablob "github.com/recruit-tech/go-ams/blob"
)

func (c *Client) newSASClient(rawurl string) (*ablob.SASClient, error) {
	return ablob.NewSASClient(rawurl,
		ablob.WithDebug(c.debug),
		ablob.WithLogger(c.logger),
		ablob.WithUserAgent(c.userAgent),
	)
}

func (c *Client) PutBlob(ctx context.Context, uploadURL *url.URL, blob io.Reader, blockID string) error {
	if ctx == nil {
		return errors.New("missing ctx")
	}
	if uploadURL == nil {
		return errors.New("missing uploadURL")
	}
	if blob == nil {
		return errors.New("missing blob")
	}
	if len(blockID) == 0 {
		return errors.New("missing blockID")
	}
	sasc, err := c.newSASClient(uploadURL.String())
	if err != nil {
		return errors.Wrap(err, "failed to construct SASClient")
	}
	return sasc.PutBlob(ctx, blob, blockID)
}

func (c *Client) PutBlockList(ctx context.Context, uploadURL *url.URL, blockList []string) error {
	if ctx == nil {
		return errors.New("missing ctx")
	}
	if uploadURL == nil {
		return errors.New("missing uploadURL")
	}
	if len(blockList) == 0 {
		return errors.New("missing blockList")
	}
	sasc, err := c.newSASClient(uploadURL.String())
	if err != nil {
		return errors.Wrap(err, "failed to construct SASClient")
	}
	return sasc.PutBlockList(ctx, blockList)
}
