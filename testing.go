package ams

import (
	"time"

	"golang.org/x/oauth2"
)

func testTokenSource() oauth2.TokenSource {
	return oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: "<<DUMMY ACCESS TOKEN>>",
		TokenType:   "Bearer",
	})
}

func testAsset() *Asset {
	return &Asset{
		ID:           "sample-id",
		State:        StateInitialized,
		Created:      formatTime(time.Now()),
		LastModified: formatTime(time.Now()),
		Name:         "Sample",
		Options:      OptionNone,
		FormatOption: FormatOptionNoFormat,
	}
}
