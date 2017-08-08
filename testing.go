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

func testAsset(id, name string) Asset {
	return Asset{
		ID:           id,
		State:        StateInitialized,
		Created:      formatTime(time.Now()),
		LastModified: formatTime(time.Now()),
		Name:         name,
		Options:      OptionNone,
		FormatOption: FormatOptionNoFormat,
	}
}
