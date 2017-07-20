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
	msVersion               = "2.15"
	dataServiceVersion      = "3.0"
	maxDataServiceVersion   = "3.0"
	requestMIMEType         = "application/json"
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

func NewClient(restAPIEndpoint, tenantDomain, clientID, clientSecret string, logger *log.Logger) (*Client, error) {
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

	parsedURL, err := url.ParseRequestURI(restAPIEndpoint)
	if err != nil {
		return nil, err
	}

	return &Client{
		URL:          parsedURL,
		httpClient:   http.DefaultClient,
		tenantDomain: tenantDomain,
		clientID:     clientID,
		clientSecret: clientSecret,
		logger:       logger,
	}, nil
}

func (c *Client) newRequest(ctx context.Context, method, spath string, body io.Reader) (*http.Request, error) {
	if len(c.credentials.AccessToken) == 0 {
		return nil, errors.New("Unauthorized, please call Auth()")
	}
	u := *c.URL
	u.Path = path.Join(c.URL.Path, spath)

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
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
	req.Header.Set("x-ms-version", msVersion)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", c.credentials.TokenType, c.credentials.AccessToken))
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
		return err
	}
	req.Header.Set("User-Agent", userAgent)

	if err := c.do(req, http.StatusOK, &c.credentials); err != nil {
		return err
	}
	return nil
}

func (c *Client) do(req *http.Request, expectedCode int, out interface{}) error {
	reqDump, err := httputil.DumpRequestOut(req, false)
	if err != nil {
		return err
	}
	c.logger.Printf("[DEBUG]\n\n%s\n", string(reqDump))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := assertStatusCode(resp, expectedCode); err != nil {
		return err
	}
	if out != nil {
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(out); err != nil {
			return err
		}
	}
	respDump, err := httputil.DumpResponse(resp, false)
	if err != nil {
		return err
	}
	c.logger.Printf("[DEBUG]\n\n%s\n", string(respDump))

	return nil
}

func encodeParams(params map[string]interface{}) (io.Reader, error) {
	encoded, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(encoded)
	return reader, nil
}

func assertStatusCode(resp *http.Response, expected int) error {
	if resp.StatusCode != expected {
		return errors.Errorf("unexpected status code, expected = %d. actual = %s %s", expected, resp.Status, resp.Request.URL.String())
	}
	return nil
}
