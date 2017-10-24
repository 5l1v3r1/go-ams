package blob

import (
	"log"
	"net/http"
)

type clientOptions struct {
	Client    *http.Client
	UserAgent string
	Logger    *log.Logger
	Debug     bool
}

type clientOption func(*clientOptions)

func WithHTTPClient(httpClient *http.Client) clientOption {
	return func(o *clientOptions) {
		o.Client = httpClient
	}
}

func WithUserAgent(userAgent string) clientOption {
	return func(o *clientOptions) {
		o.UserAgent = userAgent
	}
}

func WithLogger(logger *log.Logger) clientOption {
	return func(o *clientOptions) {
		o.Logger = logger
	}
}

func WithDebug(debug bool) clientOption {
	return func(o *clientOptions) {
		o.Debug = debug
	}
}