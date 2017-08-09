package ams

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestClient_PutBlob(t *testing.T) {
	fpath := filepath.Join("testdata", "test.mp4")
	expected, err := ioutil.ReadFile(fpath)
	if err != nil {
		t.Fatal(err)
	}

	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testRequestMethod(t, r, http.MethodPut)
		testStorageHeader(t, r.Header)

		hc := checker{
			t:          t,
			dictionary: r.Header,
		}
		hc.Assert("x-ms-blob-type", "BlockBlob")
		hc.Assert("Content-Type", "application/octet-stream")

		qc := checker{
			t:          t,
			dictionary: r.URL.Query(),
		}
		qc.Assert("comp", "block")
		qc.AssertNot("blockid", "")

		actual, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		if bytes.Compare(actual, expected) != 0 {
			t.Error("post file invalid")
		}

		if int64(len(actual)) != r.ContentLength {
			t.Errorf("unexpected ContentLength. expected: %v, actual: %v", len(actual), r.ContentLength)
		}

		w.WriteHeader(http.StatusCreated)
	})
	s := httptest.NewServer(m)
	defer s.Close()

	client := testClient(t, s.URL)
	file, err := os.Open(fpath)
	if err != nil {
		t.Fatal(err)
	}

	u, err := url.ParseRequestURI(s.URL)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("withInvalidFile", func(t *testing.T) {
		blockList, err := client.PutBlob(context.TODO(), u, nil)
		if err == nil {
			t.Error("accept invalid file")
		}
		if len(blockList) > 0 {
			t.Error("return invalid blockList")
		}
	})

	t.Run("positiveCase", func(t *testing.T) {
		blockList, err := client.PutBlob(context.TODO(), u, file)
		if err != nil {
			t.Error(err)
		}
		if len(blockList) == 0 {
			t.Error("return empty blockList")
		}
	})
}

func TestClient_PutBlockList(t *testing.T) {
	blockList := []int{1}

	expected, err := buildBlockListXML(blockList)
	if err != nil {
		t.Fatal(err)
	}

	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testRequestMethod(t, r, http.MethodPut)
		testStorageHeader(t, r.Header)

		qc := checker{
			t:          t,
			dictionary: r.URL.Query(),
		}
		qc.Assert("comp", "blocklist")

		if r.ContentLength != int64(len(expected)) {
			t.Errorf("unexpected ContentLength. expected: %v, actual: %v", len(expected), r.ContentLength)
		}

		actual, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}

		if bytes.Compare(actual, expected) != 0 {
			t.Error("post invalid file")
		}

		w.WriteHeader(http.StatusCreated)
	})
	s := httptest.NewServer(m)
	client := testClient(t, s.URL)

	u, err := url.ParseRequestURI(s.URL)
	if err != nil {
		t.Fatal(err)
	}

	if err := client.PutBlockList(context.TODO(), u, blockList); err != nil {
		t.Error(err)
	}
}
