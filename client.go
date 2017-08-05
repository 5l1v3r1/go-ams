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
	"path"
	"runtime"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

const (
	Resource              = "https://rest.media.azure.net"
	APIVersion            = "2.15"
	StorageAPIVersion     = "2017-04-17"
	DataServiceVersion    = "3.0"
	MaxDataServiceVersion = "3.0"
	version               = "0.2.0"
)

var (
	userAgent = fmt.Sprintf("Go/%s (%s-%s) go-ams/%s", runtime.Version(), runtime.GOARCH, runtime.GOOS, version)
)

type Client struct {
	baseURL     *url.URL
	tokenSource oauth2.TokenSource

	httpClient *http.Client

	logger *log.Logger
	debug  bool
}

func NewClient(urlStr string, tokenSource oauth2.TokenSource) (*Client, error) {
	if tokenSource == nil {
		return nil, errors.New("missing tokenSource")
	}
	u, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, errors.Wrapf(err, "url parse failed: %s", urlStr)
	}
	defaultLogger := log.New(ioutil.Discard, "", log.LstdFlags)
	return &Client{
		baseURL:     u,
		tokenSource: tokenSource,

		httpClient: http.DefaultClient,

		logger: defaultLogger,
		debug:  false,
	}, nil
}

func (c *Client) SetDebug(debug bool) {
	c.debug = debug
}

func (c *Client) SetLogger(logger *log.Logger) {
	c.logger = logger
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

	token, err := c.tokenSource.Token()
	if err != nil {
		return nil, errors.Wrapf(err, "get access token failed")
	}
	token.SetAuthHeader(req)

	return req, nil
}

func (c *Client) newStorageRequest(ctx context.Context, method string, u url.URL, opts ...requestOption) (*http.Request, error) {
	option := defaultStorageRequestOption()
	return c.newCommonRequest(ctx, &u, method, option, opts...)
}

func (c *Client) do(req *http.Request, expectedCode int, out interface{}) error {
	if c.debug {
		reqDump, err := httputil.DumpRequestOut(req, false)
		if err != nil {
			return errors.Wrap(err, "request dump failed")
		}
		c.logger.Printf("[DEBUG] url = %s", req.URL.String())
		c.logger.Printf("[DEBUG] request header\n%s", string(reqDump))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var body io.Reader
	body = resp.Body

	if c.debug {
		respDump, err := httputil.DumpResponse(resp, false)
		if err != nil {
			return errors.Wrap(err, "response dump failed")
		}
		c.logger.Printf("[DEBUG] url = %s", req.URL.String())
		c.logger.Printf("[DEBUG] response header\n%s", string(respDump))

		var b bytes.Buffer
		if _, err := b.ReadFrom(body); err != nil {
			return errors.Wrap(err, "response body read failed")
		}
		c.logger.Printf("[DEBUG] body\n%s", b.String())

		body = &b
	}

	if err := assertStatusCode(resp, expectedCode); err != nil {
		return err
	}

	if out != nil {
		decoder := json.NewDecoder(body)
		if err := decoder.Decode(out); err != nil {
			return errors.Wrap(err, "response decode failed")
		}

		if c.debug {
			c.logger.Printf("[DEBUG] parsed body\n%#v", out)
		}
	}

	return nil
}

func (c *Client) buildURI(spath string) string {
	u := *c.baseURL
	u.Path = path.Join(u.Path, spath)
	return u.String()
}
