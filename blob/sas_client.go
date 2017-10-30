package blob

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/orisano/httpc"
	"github.com/pkg/errors"
)

const (
	APIVersion = "2017-04-17"
)

var DefaultUserAgent = fmt.Sprintf("Go/%s (%s-%s) go-ams.blob", runtime.Version(), runtime.GOARCH, runtime.GOOS)

type SASClient struct {
	rb         *httpc.RequestBuilder
	httpClient *http.Client

	userAgent string
	logger    *log.Logger
	debug     bool

	TimeNow func() time.Time
}

func NewSASClient(rawurl string, opts ...clientOption) (*SASClient, error) {
	options := &clientOptions{
		Client:    http.DefaultClient,
		UserAgent: DefaultUserAgent,
		Logger:    log.New(ioutil.Discard, "", log.Lshortfile),
		Debug:     false,
	}
	for _, opt := range opts {
		opt(options)
	}

	h := make(http.Header)
	h.Set("x-ms-version", APIVersion)
	h.Set("User-Agent", options.UserAgent)

	rb, err := httpc.NewRequestBuilder(rawurl, h)
	if err != nil {
		return nil, err
	}

	if options.Debug {
		httpc.InjectDebugTransport(options.Client, os.Stderr)
	}

	return &SASClient{
		rb:         rb,
		httpClient: options.Client,
		userAgent:  options.UserAgent,
		logger:     options.Logger,
		debug:      options.Debug,
		TimeNow:    time.Now,
	}, nil
}

func withDate(t time.Time) httpc.RequestOption {
	return func(o *httpc.RequestOptions) error {
		o.Header.Set("Date", t.UTC().Format(time.RFC3339))
		return nil
	}
}

func (c *SASClient) PutBlob(ctx context.Context, blob io.Reader, blockID string) error {
	if ctx == nil {
		return errors.New("missing ctx")
	}
	if blob == nil {
		return errors.New("missing blob")
	}
	if len(blockID) == 0 {
		return errors.New("missing blockID")
	}
	req, err := c.rb.NewRequest(ctx, http.MethodPut, "",
		withDate(c.TimeNow()),
		httpc.SetHeaderField("x-ms-blob-type", "BlockBlob"),
		httpc.AddQuery("comp", "block"),
		httpc.AddQuery("blockid", blockID),
		httpc.WithBinary(blob),
		httpc.EnforceContentLength,
	)
	if err != nil {
		return errors.Wrap(err, "failed to construct http request")
	}
	c.logger.Print("[INFO] put blob ...")
	resp, err := httpc.Retry(c.httpClient, req)
	if err != nil {
		return errors.Wrap(err, "failed to http request")
	}
	defer resp.Body.Close()

	if got := resp.StatusCode; got != http.StatusCreated {
		return errors.Errorf("unexpected status code. expected: %v, but got: %v", http.StatusCreated, got)
	}

	c.logger.Print("[INFO] completed")
	return nil
}

func (c *SASClient) PutBlockList(ctx context.Context, blockList []string) error {
	if ctx == nil {
		return errors.New("missing ctx")
	}
	if len(blockList) == 0 {
		return errors.New("missing blockList")
	}
	req, err := c.rb.NewRequest(ctx, http.MethodPut, "",
		withDate(c.TimeNow()),
		httpc.AddQuery("comp", "blocklist"),
		httpc.WithXML(&BlockList{Blocks: blockList}),
		httpc.EnforceContentLength,
	)
	if err != nil {
		return errors.Wrap(err, "failed to construct http request")
	}
	c.logger.Print("[INFO] put block list ...")
	resp, err := httpc.Retry(c.httpClient, req)
	if err != nil {
		return errors.Wrap(err, "failed to http request")
	}
	defer resp.Body.Close()

	if got := resp.StatusCode; got != http.StatusCreated {
		return errors.Errorf("unexpected status code. expected: %v, but got: %v", http.StatusCreated, got)
	}

	c.logger.Print("[INFO] completed")
	return nil
}
