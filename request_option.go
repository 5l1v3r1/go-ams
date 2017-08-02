package ams

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type requestOptions struct {
	Body    io.Reader
	BaseURL url.URL
	Header  http.Header
	Params  url.Values
}

type requestOption func(*requestOptions) error

func composeOptions(opts ...requestOption) requestOption {
	return func(option *requestOptions) error {
		for _, opt := range opts {
			if err := opt(option); err != nil {
				return err
			}
		}
		return nil
	}
}

func defaultRequestOption(client *Client) *requestOptions {
	option := &requestOptions{
		Header: http.Header{},
		Params: url.Values{},
	}

	option.BaseURL = *client.baseURL
	option.Header.Set("User-Agent", userAgent)

	return option
}

func useAMS(client *Client) requestOption {
	return composeOptions(
		withOData(false),
		withDataServiceVersion(),
		setAPIVersion(apiVersion),
	)
}

func useStorageAPI() requestOption {
	return composeOptions(
		withDateHeader(),
		setAPIVersion(storageAPIVersion),
	)
}

func withDataServiceVersion() requestOption {
	return func(option *requestOptions) error {
		option.Header.Set("DataServiceVersion", dataServiceVersion)
		option.Header.Set("MaxDataServiceVersion", maxDataServiceVersion)
		return nil
	}
}

func withQuery(params url.Values) requestOption {
	return func(option *requestOptions) error {
		for k, vs := range params {
			for _, v := range vs {
				option.Params.Add(k, v)
			}
		}
		return nil
	}
}

func withForm(params url.Values) requestOption {
	return func(option *requestOptions) error {
		option.Body = strings.NewReader(params.Encode())
		return nil
	}
}

func withBody(body io.Reader) requestOption {
	return func(option *requestOptions) error {
		option.Body = body
		return nil
	}
}

func withJSON(data interface{}) requestOption {
	return func(option *requestOptions) error {
		encoded, err := json.Marshal(data)
		if err != nil {
			return errors.Wrap(err, "json marshal failed")
		}
		option.Body = bytes.NewReader(encoded)
		return nil
	}
}

func withBytes(b []byte) requestOption {
	return func(option *requestOptions) error {
		option.Body = bytes.NewReader(b)
		return nil
	}
}

func setURL(rawurl string) requestOption {
	return func(option *requestOptions) error {
		u, err := url.ParseRequestURI(rawurl)
		if err != nil {
			return errors.Wrapf(err, "url parse failed: %s", rawurl)
		}
		option.BaseURL = *u
		return nil
	}
}

func withDateHeader() requestOption {
	return func(option *requestOptions) error {
		option.Header.Set("Date", time.Now().UTC().Format(time.RFC3339))
		return nil
	}
}

func withBlobType(blobType string) requestOption {
	return func(option *requestOptions) error {
		option.Header.Set("x-ms-blob-type", blobType)
		return nil
	}
}

func withContentType(mimeType string) requestOption {
	return func(option *requestOptions) error {
		option.Header.Set("Content-Type", mimeType)
		return nil
	}
}

func setAccept(mimeType string) requestOption {
	return func(option *requestOptions) error {
		option.Header.Set("Accept", mimeType)
		return nil
	}
}

func setAPIVersion(version string) requestOption {
	return func(option *requestOptions) error {
		option.Header.Set("x-ms-version", version)
		return nil
	}
}

func withOData(verbose bool) requestOption {
	contentType := requestMIMEType
	accept := responseMIMEType
	if verbose {
		contentType += ";odata=verbose"
		accept += ";odata=verbose"
	}
	return composeOptions(withContentType(contentType), setAccept(accept))
}
