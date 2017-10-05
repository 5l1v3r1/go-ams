package ams

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"runtime"

	"github.com/k0kubun/pp"
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
	UserAgent *string
	Logger    *log.Logger
	Debug     bool
}

type clientOption func(*clientOptions)

func SetUserAgent(userAgent string) clientOption {
	return func(options *clientOptions) {
		options.UserAgent = &userAgent
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
	baseURL *url.URL

	authorizedClient *http.Client

	userAgent string
	logger    *log.Logger
	debug     bool
}

func NewClient(urlStr string, authorizedClient *http.Client, opts ...clientOption) (*Client, error) {
	if authorizedClient == nil {
		return nil, errors.New("missing authorizedClient")
	}
	u, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, errors.Wrapf(err, "url parse failed: %s", urlStr)
	}

	globalDebug := len(os.Getenv("GO_AMS_DEBUG")) != 0

	var options clientOptions
	for _, opt := range opts {
		opt(&options)
	}
	logger := options.Logger
	if logger == nil {
		if globalDebug {
			logger = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)
		} else {
			logger = log.New(ioutil.Discard, "", log.LstdFlags|log.Lshortfile)
		}
	}
	userAgent := options.UserAgent
	if userAgent == nil {
		userAgent = &defaultUserAgent
	}

	return &Client{
		baseURL: u,

		authorizedClient: authorizedClient,

		userAgent: *userAgent,
		logger:    logger,
		debug:     options.Debug || globalDebug,
	}, nil
}

func (c *Client) newCommonRequest(ctx context.Context, u *url.URL, method string, option *requestOptions, opts ...requestOption) (*http.Request, error) {
	if err := composeOptions(opts...)(option); err != nil {
		return nil, errors.Wrap(err, "option apply failed")
	}
	if len(option.Params) != 0 {
		q := u.Query()
		mergeValues(q, option.Params)
		u.RawQuery = q.Encode()
	}
	req, err := http.NewRequest(method, u.String(), option.Body)
	if err != nil {
		return nil, err
	}
	req.Header = option.Header
	req.Header.Set("User-Agent", c.userAgent)

	req = req.WithContext(ctx)
	return req, nil
}

func (c *Client) newRequest(ctx context.Context, method, spath string, opts ...requestOption) (*http.Request, error) {
	option := defaultRequestOption()
	u := *c.baseURL
	u.Path = path.Join(u.Path, spath)

	req, err := c.newCommonRequest(ctx, &u, method, option, opts...)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) newStorageRequest(ctx context.Context, method string, u url.URL, opts ...requestOption) (*http.Request, error) {
	option := defaultStorageRequestOption()
	return c.newCommonRequest(ctx, &u, method, option, opts...)
}

func (c *Client) doWithClient(httpClient *http.Client, req *http.Request, expectedCode int, out interface{}) error {
	if c.debug {
		dump, err := httputil.DumpRequestOut(req, false)
		if err != nil {
			return errors.Wrap(err, "request dump failed")
		}
		c.logger.Print("[DEBUG] request header\n" + string(dump))

		if req.Body != nil {
			var b []byte
			req.Body, b, err = sniffBody(req.Body)
			if err != nil {
				return errors.Wrap(err, "request body read failed")
			}
			c.logger.Print("[DEBUG] request body\n" + string(b))
		}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if c.debug {
		dump, err := httputil.DumpResponse(resp, false)
		if err != nil {
			return errors.Wrap(err, "response dump failed")
		}
		c.logger.Print("[DEBUG] response header\n" + string(dump))

		var b []byte
		resp.Body, b, err = sniffBody(resp.Body)
		if err != nil {
			return errors.Wrap(err, "response body read failed")
		}
		c.logger.Print("[DEBUG] response body\n" + string(b))
	}

	if err := assertStatusCode(resp, expectedCode); err != nil {
		return err
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return errors.Wrap(err, "response decode failed")
		}

		if c.debug {
			c.logger.Print("[DEBUG] parsed body")
			c.logger.Print(pp.Sprint(out))
		}
	}

	return nil
}

func (c *Client) do(req *http.Request, expectedCode int, out interface{}) error {
	return c.doWithClient(c.authorizedClient, req, expectedCode, out)
}

func (c *Client) get(ctx context.Context, spath string, out interface{}) error {
	req, err := c.newRequest(ctx, http.MethodGet, spath)
	if err != nil {
		return errors.Wrap(err, "request construct failed")
	}
	if err := c.do(req, http.StatusOK, out); err != nil {
		return errors.Wrap(err, "request failed")
	}
	return nil
}

func (c *Client) post(ctx context.Context, spath string, in interface{}, out interface{}, opts ...requestOption) error {
	opts = append(opts, withJSON(in))
	req, err := c.newRequest(ctx, http.MethodPost, spath, opts...)
	if err != nil {
		return errors.Wrap(err, "request construct failed")
	}
	if err := c.do(req, http.StatusCreated, out); err != nil {
		return errors.Wrap(err, "request failed")
	}
	return nil
}

func (c *Client) buildURI(spath string) string {
	u := *c.baseURL
	u.Path = path.Join(u.Path, spath)
	return u.String()
}

func sniffBody(r io.ReadCloser) (io.ReadCloser, []byte, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return r, []byte{}, err
	}
	return ioutil.NopCloser(bytes.NewReader(b)), b, nil
}
