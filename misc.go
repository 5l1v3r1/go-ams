package ams

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

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
