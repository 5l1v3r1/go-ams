package ams

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

func assertStatusCode(resp *http.Response, expected int) error {
	if resp.StatusCode != expected {
		return errors.Errorf("unexpected status code, expected = %d, actual = %s <= %s", expected, resp.Status, resp.Request.URL.String())
	}
	return nil
}

func toResource(name, id string) string {
	return fmt.Sprintf("%s('%s')", name, id)
}

func mergeValues(a, b url.Values) {
	for k, vs := range b {
		for _, v := range vs {
			a.Set(k, v)
		}
	}
}
