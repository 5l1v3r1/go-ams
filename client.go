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
)

const (
	Resource              = "https://rest.media.azure.net"
	version               = "0.1.0"
	apiVersion            = "2.15"
	storageAPIVersion     = "2017-04-17"
	dataServiceVersion    = "3.0"
	maxDataServiceVersion = "3.0"
	requestMIMEType       = "application/json"
	responseMIMEType      = "application/json"
)

var (
	userAgent = fmt.Sprintf("Go/%s (%s-%s) go-ams/%s", runtime.Version(), runtime.GOARCH, runtime.GOOS, version)
)

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client

	logger *log.Logger

	debug bool
}

func NewClient(httpClient *http.Client, urlStr string, logger *log.Logger) (*Client, error) {
	if logger == nil {
		logger = log.New(ioutil.Discard, "", log.LstdFlags)
	}

	u, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, errors.Wrapf(err, "url parse failed: %s", urlStr)
	}

	return &Client{
		baseURL:    u,
		httpClient: httpClient,
		logger:     logger,
		debug:      false,
	}, nil
}

func (c *Client) SetDebug(debug bool) {
	c.debug = debug
}

func (c *Client) newRequest(ctx context.Context, method, spath string, opts ...requestOption) (*http.Request, error) {
	option := defaultRequestOption(c)
	if err := composeOptions(opts...)(option); err != nil {
		return nil, errors.Wrap(err, "option apply failed")
	}

	u := option.BaseURL
	u.Path = path.Join(u.Path, spath)
	if len(option.Params) != 0 {
		q := u.Query()
		for k, vs := range option.Params {
			for _, v := range vs {
				q.Add(k, v)
			}
		}
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequest(method, u.String(), option.Body)
	if err != nil {
		return nil, errors.Wrap(err, "request build failed")
	}
	req.Header = option.Header

	req = req.WithContext(ctx)
	return req, nil
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
		return errors.Wrap(err, "request failed")
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
