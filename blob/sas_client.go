package blob

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/orisano/httpc"
	"github.com/pkg/errors"
)

const (
	APIVersion = "2017-04-17"
)

var TimeNow func() time.Time = time.Now
var DefaultUserAgent = fmt.Sprintf("Go/%s (%s-%s) go-ams.blob", runtime.Version(), runtime.GOARCH, runtime.GOOS)

type SASClient struct {
	u          *url.URL
	httpClient *http.Client

	userAgent string
	logger    *log.Logger
	debug     bool
}

func NewSASClient(rawurl string, opts ...clientOption) (*SASClient, error) {
	if len(rawurl) == 0 {
		return nil, errors.New("missing rawurl")
	}
	u, err := url.ParseRequestURI(rawurl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse url")
	}

	options := &clientOptions{
		Client:    http.DefaultClient,
		UserAgent: DefaultUserAgent,
		Logger:    log.New(ioutil.Discard, "", log.Lshortfile),
		Debug:     false,
	}
	for _, opt := range opts {
		opt(options)
	}

	return &SASClient{
		u:          u,
		httpClient: options.Client,
		userAgent:  options.UserAgent,
		logger:     options.Logger,
		debug:      options.Debug,
	}, nil
}

func (c *SASClient) defaultHeader() http.Header {
	h := http.Header{}
	h.Set("x-ms-version", APIVersion)
	h.Set("Date", TimeNow().UTC().Format(time.RFC3339))
	h.Set("User-Agent", c.userAgent)
	return h
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
	header := c.defaultHeader()
	header.Set("x-ms-blob-type", "BlockBlob")

	params := url.Values{}
	params.Set("comp", "block")
	params.Set("blockid", blockID)

	req, err := httpc.NewRequest(ctx, http.MethodPut, c.u.String(),
		httpc.WithHeader(header),
		httpc.WithQueries(params),
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
	req, err := httpc.NewRequest(ctx, http.MethodPut, c.u.String(),
		httpc.WithHeader(c.defaultHeader()),
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
	c.logger.Print("[INFO] completed")
	return nil
}
