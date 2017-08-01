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
	"strings"

	"github.com/pkg/errors"
)

const (
	azureADAuthServerFormat = "https://login.microsoftonline.com/%s/oauth2/token"
	resource                = "https://rest.media.azure.net"
	grantType               = "client_credentials"
	version                 = "0.1.0"
	amsAPIVersion           = "2.15"
	dataServiceVersion      = "3.0"
	maxDataServiceVersion   = "3.0"
	requestMIMEType         = "application/json"
	responseMIMEType        = "application/json"
)

var (
	userAgent = fmt.Sprintf("Go/%s (%s-%s) go-ams/%s",
		runtime.Version(),
		runtime.GOARCH,
		runtime.GOOS,
		version,
	)
)

type Client struct {
	URL        *url.URL
	httpClient *http.Client

	tenantDomain           string
	clientID, clientSecret string

	logger *log.Logger

	credentials Credentials

	debug bool
}

type Credentials struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	ExpiresOn    string `json:"expires_on"`
	ExtExpiresIn string `json:"ext_expires_in"`
	NotBefore    string `json:"not_before"`
	Resource     string `json:"resource"`
	TokenType    string `json:"token_type"`
}

func (c *Credentials) Token() string {
	return fmt.Sprintf("%s %s", c.TokenType, c.AccessToken)
}

func NewClient(apiEndpoint, tenantDomain, clientID, clientSecret string, logger *log.Logger) (*Client, error) {
	if len(tenantDomain) == 0 {
		return nil, errors.New("missing tenantDomain")
	}
	if len(clientID) == 0 {
		return nil, errors.New("missing clientID")
	}
	if len(clientSecret) == 0 {
		return nil, errors.New("missing clientSecret")
	}
	if logger == nil {
		logger = log.New(ioutil.Discard, "", log.LstdFlags)
	}

	parsedURL, err := url.ParseRequestURI(apiEndpoint)
	if err != nil {
		return nil, errors.Wrapf(err, "url parse failed: %s", apiEndpoint)
	}

	return &Client{
		URL:          parsedURL,
		httpClient:   http.DefaultClient,
		tenantDomain: tenantDomain,
		clientID:     clientID,
		clientSecret: clientSecret,
		logger:       logger,
		debug:        false,
	}, nil
}

func (c *Client) SetDebug(debug bool) {
	c.debug = debug
}

func (c *Client) newRequest(ctx context.Context, method, spath string, body io.Reader) (*http.Request, error) {
	if len(c.credentials.AccessToken) == 0 {
		return nil, errors.New("no access token")
	}
	u := *c.URL
	u.Path = path.Join(c.URL.Path, spath)

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, errors.Wrap(err, "request build failed")
	}
	c.setDefaultHeader(req)
	req.Header.Set("Content-Type", requestMIMEType)
	req.Header.Set("Accept", requestMIMEType)
	req.Header.Set("DataServiceVersion", dataServiceVersion)
	req.Header.Set("MaxDataServiceVersion", maxDataServiceVersion)

	req = req.WithContext(ctx)

	return req, nil
}

func (c *Client) setDefaultHeader(req *http.Request) {
	req.Header.Set("x-ms-version", amsAPIVersion)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Authorization", c.credentials.Token())
}

func (c *Client) Auth() error {
	authURL := fmt.Sprintf(azureADAuthServerFormat, c.tenantDomain)

	params := url.Values{}
	params.Add("grant_type", grantType)
	params.Add("client_id", c.clientID)
	params.Add("client_secret", c.clientSecret)
	params.Add("resource", resource)
	body := strings.NewReader(params.Encode())

	req, err := http.NewRequest(http.MethodPost, authURL, body)
	if err != nil {
		return errors.Wrap(err, "auth request build failed")
	}
	req.Header.Set("User-Agent", userAgent)

	if err := c.do(req, http.StatusOK, &c.credentials); err != nil {
		return errors.Wrap(err, "auth request failed")
	}
	return nil
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
	u := *c.URL
	u.Path = path.Join(u.Path, spath)
	return u.String()
}

func encodeParams(params interface{}) (io.Reader, error) {
	encoded, err := json.Marshal(params)
	if err != nil {
		return nil, errors.Wrap(err, "json marshal failed")
	}
	reader := bytes.NewReader(encoded)
	return reader, nil
}

func assertStatusCode(resp *http.Response, expected int) error {
	if resp.StatusCode != expected {
		return errors.Errorf("unexpected status code, expected = %d, actual = %s <= %s", expected, resp.Status, resp.Request.URL.String())
	}
	return nil
}

func toResource(name, id string) string {
	return fmt.Sprintf("%s('%s')", name, id)
}
