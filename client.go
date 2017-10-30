package ams

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"

	"github.com/orisano/httpc"
	"github.com/pkg/errors"
)

const (
	Resource              = "https://rest.media.azure.net"
	APIVersion            = "2.15"
	StorageAPIVersion     = "2017-04-17"
	DataServiceVersion    = "3.0"
	MaxDataServiceVersion = "3.0"
	version               = "0.3.0"
)

var (
	defaultUserAgent = fmt.Sprintf("Go/%s (%s-%s) go-ams/%s", runtime.Version(), runtime.GOARCH, runtime.GOOS, version)
)

type clientOptions struct {
	UserAgent string
	Logger    *log.Logger
	Debug     bool
}

type clientOption func(*clientOptions)

func SetUserAgent(userAgent string) clientOption {
	return func(options *clientOptions) {
		options.UserAgent = userAgent
	}
}

func SetLogger(logger *log.Logger) clientOption {
	return func(options *clientOptions) {
		options.Logger = logger
	}
}

func SetDebug(debug bool) clientOption {
	return func(options *clientOptions) {
		options.Debug = debug
	}
}

type Client struct {
	rb *httpc.RequestBuilder

	authorizedClient *http.Client

	userAgent string
	logger    *log.Logger
	debug     bool
}

func NewClient(urlStr string, authorizedClient *http.Client, opts ...clientOption) (*Client, error) {
	if authorizedClient == nil {
		return nil, errors.New("missing authorizedClient")
	}

	options := &clientOptions{
		UserAgent: defaultUserAgent,
	}
	for _, opt := range opts {
		opt(options)
	}

	h := make(http.Header)
	h.Set("x-ms-version", APIVersion)
	h.Set("DataServiceVersion", DataServiceVersion)
	h.Set("MaxDataServiceVersion", MaxDataServiceVersion)
	h.Set("UserAgent", options.UserAgent)
	h.Set("Accept", "application/json")

	rb, err := httpc.NewRequestBuilder(urlStr, h)
	if err != nil {
		return nil, err
	}

	globalDebug := len(os.Getenv("GO_AMS_DEBUG")) != 0
	debug := options.Debug || globalDebug

	logger := options.Logger
	if logger == nil {
		if debug {
			logger = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)
		} else {
			logger = log.New(ioutil.Discard, "", log.LstdFlags|log.Lshortfile)
		}
	}

	if debug {
		httpc.InjectDebugTransport(authorizedClient, os.Stderr)
	}

	return &Client{
		rb: rb,

		authorizedClient: authorizedClient,

		userAgent: options.UserAgent,
		logger:    logger,
		debug:     debug,
	}, nil
}

func (c *Client) newRequest(ctx context.Context, method, spath string, opts ...httpc.RequestOption) (*http.Request, error) {
	return c.rb.NewRequest(ctx, method, spath, opts...)
}

func (c *Client) do(req *http.Request, expectedCode int, out interface{}) error {
	resp, err := httpc.Retry(c.authorizedClient, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if got := resp.StatusCode; got != expectedCode {
		return errors.Errorf("unexpected status code. expected: %v, but got: %v", expectedCode, got)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return errors.Wrap(err, "failed to decode response")
		}
	}
	return nil
}

func (c *Client) get(ctx context.Context, spath string, out interface{}) error {
	req, err := c.newRequest(ctx, http.MethodGet, spath)
	if err != nil {
		return errors.Wrap(err, "failed to construct GET request")
	}
	if err := c.do(req, http.StatusOK, out); err != nil {
		return errors.Wrap(err, "failed to GET request")
	}
	return nil
}

func (c *Client) post(ctx context.Context, spath string, in interface{}, out interface{}, opts ...httpc.RequestOption) error {
	opts = append(opts, httpc.WithJSON(in))
	req, err := c.newRequest(ctx, http.MethodPost, spath, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to construct POST request")
	}
	if err := c.do(req, http.StatusCreated, out); err != nil {
		return errors.Wrap(err, "failed to POST request")
	}
	return nil
}

func (c *Client) buildURI(spath string) string {
	u := *c.rb.BaseURL()
	u.Path = path.Join(u.Path, spath)
	return u.String()
}
