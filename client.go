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
	"net/url"
	"os"
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
	requestMIMEtype         = "application/json"
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

type AssetOption int

const (
	OptionNone                        = 0
	OptionStorageEncrypted            = 1
	OptionCommonEncryptionProtected   = 2
	OptionEnvelopeEncryptionProtected = 4
)

type Asset struct {
	ID                 string `json:"Id"`
	State              int    `json:"State"`
	Created            string `json:"Created"`
	LastModified       string `json:"LastModified"`
	Name               string `json:"Name"`
	Options            int    `json:"Options"`
	FormatOption       int    `json:"FormatOption"`
	URI                string `json:"Uri"`
	StorageAccountName string `json:"StorageAccountName"`
}

type AssetFile struct {
	ID              string `json:"Id"`
	Name            string `json:"Name"`
	ContentFileSize string `json:"ContentFileSize"`
	ParentAssetId   string `json:"ParentAssetId"`
	IsPrimary       bool   `json:"IsPrimary"`
	LastModified    string `json:"LastModified"`
	Created         string `json:"Created"`
	MIMEType        string `json:"MimeType"`
	ContentChecksum string `json:"ContentChecksum"`
}

type AccessPolicy struct {
	ID                string  `json:"Id"`
	Created           string  `json:"Created"`
	LastModified      string  `json:"LastModified"`
	Name              string  `json:"Name"`
	DurationInMinutes float64 `json:"DurationInMinutes"`
	Permissions       int     `json:"Permissions"`
}

type Locator struct {
	ID                     string `json:"Id"`
	ExpirationDateTime     string `json:"ExpirationDateTime"`
	Type                   int    `json:"Type"`
	Path                   string `json:"Path"`
	BaseURI                string `json:"BaseUri"`
	ContentAccessComponent string `json:"ContentAccessComponent"`
	AccessPolicyID         string `json:"AccessPolicyId"`
	AssetID                string `json:"AssetID"`
	StartTime              string `json:"StartTime"`
	Name                   string `json:"Name"`
}

// TYPE:

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
	req.Header.Set("Content-Type", requestMIMEtype)
	req.Header.Set("Accept", requestMIMEtype)
	req.Header.Set("x-ms-version", msVersion)
	req.Header.Set("DataServiceVersion", dataServiceVersion)
	req.Header.Set("MaxDataServiceVersion", maxDataServiceVersion)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", c.credentials.TokenType, c.credentials.AccessToken))

	req = req.WithContext(ctx)

	return req, nil
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

func (c *Client) GetAssets() ([]Asset, error) {
	return c.GetAssetsWithContext(context.Background())
}

func (c *Client) GetAssetsWithContext(ctx context.Context) ([]Asset, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "Assets", nil)
	if err != nil {
		return nil, err
	}
	var out struct {
		Assets []Asset `json:"value"`
	}
	if err := c.do(req, http.StatusOK, &out); err != nil {
		return nil, err
	}
	return out.Assets, nil
}

func (c *Client) CreateAsset(name string) (*Asset, error) {
	return c.CreateAssetWithContext(context.Background(), name)
}

func (c *Client) CreateAssetWithContext(ctx context.Context, name string) (*Asset, error) {
	params := map[string]interface{}{
		"Name": name,
	}
	body, err := encodeParams(params)
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest(ctx, http.MethodPost, "Assets", body)
	if err != nil {
		return nil, err
	}
	var out Asset
	if err := c.do(req, http.StatusCreated, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateAssetFile(assetID, name, mimeType string) (*AssetFile, error) {
	return c.CreateAssetFileWithContext(context.Background(), assetID, name, mimeType)
}

func (c *Client) CreateAssetFileWithContext(ctx context.Context, assetID, name, mimeType string) (*AssetFile, error) {
	params := map[string]interface{}{
		"IsEncrypted":   "false",
		"IsPrimary":     "false",
		"MimeType":      mimeType,
		"Name":          name,
		"ParentAssetId": assetID,
	}
	body, err := encodeParams(params)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest(ctx, http.MethodPost, "Files", body)
	if err != nil {
		return nil, err
	}
	var out AssetFile
	if err := c.do(req, http.StatusCreated, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateAccessPolicy(name, durationInMinutes, permissions string) (*AccessPolicy, error) {
	return c.CreateAccessPolicyWithContext(context.Background(), name, durationInMinutes, permissions)
}

func (c *Client) CreateAccessPolicyWithContext(ctx context.Context, name, durationInMinutes, permissions string) (*AccessPolicy, error) {
	params := map[string]interface{}{
		"Name":              name,
		"DurationInMinutes": durationInMinutes,
		"Permissions":       permissions,
	}
	body, err := encodeParams(params)
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest(ctx, http.MethodPost, "AccessPolicies", body)
	if err != nil {
		return nil, err
	}
	var out AccessPolicy
	if err := c.do(req, http.StatusCreated, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateLocator(accessPolicyID, assetID, startTime string, locatorType int) (*Locator, error) {
	return c.CreateLocatorWithContext(context.Background(), accessPolicyID, assetID, startTime, locatorType)
}

func (c *Client) CreateLocatorWithContext(ctx context.Context, accessPolicyID, assetID, startTime string, locatorType int) (*Locator, error) {
	params := map[string]interface{}{
		"AccessPolicyId": accessPolicyID,
		"AssetId":        assetID,
		"StartTime":      startTime,
		"Type":           locatorType,
	}
	body, err := encodeParams(params)
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest(ctx, http.MethodPost, "AccessPolicies", body)
	if err != nil {
		return nil, err
	}
	var out Locator
	if err := c.do(req, http.StatusCreated, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// API:

func (c *Client) do(req *http.Request, expectedCode int, out interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := assertStatusCode(resp, expectedCode); err != nil {
		return err
	}
	r := io.TeeReader(resp.Body, os.Stdout)
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(out); err != nil {
		return err
	}
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
