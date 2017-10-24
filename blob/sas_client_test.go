package blob

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestClient_PutBlob(t *testing.T) {
	fpath := filepath.Join("testdata", "test.mp4")
	expected, err := ioutil.ReadFile(fpath)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	TimeNow = func() time.Time { return now }

	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("unexpected http method. expected: %v, got: %v", http.MethodPut, r.Method)
		}
		h := r.Header
		if got := h.Get("x-ms-version"); got != APIVersion {
			t.Errorf("unexpected x-ms-version header. expected: %v, got: %v", APIVersion, got)
		}
		if got := h.Get("Date"); got != now.UTC().Format(time.RFC3339) {
			t.Errorf("unexpected Date header. expected: %v, got: %v", now.UTC().Format(time.RFC3339), got)
		}
		if got := h.Get("User-Agent"); got != DefaultUserAgent {
			t.Errorf("unexpected User-Agent. expected: %v, got: %v", DefaultUserAgent, got)
		}
		if got := h.Get("x-ms-blob-type"); got != "BlockBlob" {
			t.Errorf("unexpected x-ms-blob-type. expected: BlockBlob, got: %v", got)
		}
		if got := h.Get("Content-Type"); got != "application/octet-stream" {
			t.Errorf("unexpected Content-Type. expected: application/octet-stream, got: %v", got)
		}
		q := r.URL.Query()
		if got := q.Get("comp"); got != "block" {
			t.Errorf("unexpected comp params. expected: block, got: %v", got)
		}
		if got := q.Get("blockid"); got == "" {
			t.Errorf("unexpected blockid params. expected: not empty, got: empty")
		}

		got, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		if bytes.Compare(got, expected) != 0 {
			t.Error("post file unexpected")
		}

		if int64(len(got)) != r.ContentLength {
			t.Errorf("unexpected ContentLength. expected: %v, actual: %v", len(got), r.ContentLength)
		}

		w.WriteHeader(http.StatusCreated)
	})

	s := httptest.NewServer(m)
	defer s.Close()

	client, err := NewSASClient(s.URL)
	if err != nil {
		t.Fatal(err)
	}

	file, err := os.Open(fpath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	b := new(bytes.Buffer)
	io.Copy(b, file)

	t.Run("withInvalidBlob", func(t *testing.T) {
		err := client.PutBlob(context.TODO(), nil, "test-block-id")
		if err == nil {
			t.Error("accept invalid blob")
		}
	})
	t.Run("withInvalidBlockID", func(t *testing.T) {
		err := client.PutBlob(context.TODO(), b, "")
		if err == nil {
			t.Error("accept invalid blockID")
		}
	})
	t.Run("positiveCase", func(t *testing.T) {
		err := client.PutBlob(context.TODO(), b, "test-block-id")
		if err != nil {
			t.Error(err)
		}
	})
}

func TestClient_PutBlockList(t *testing.T) {
	blockList := []string{"sample-block-id-1", "sample-block-id-2"}
	expected := []byte(xml.Header + `<BlockList><Latest>sample-block-id-1</Latest><Latest>sample-block-id-2</Latest></BlockList>`)

	now := time.Now()
	TimeNow = func() time.Time { return now }

	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("unexpected http method. expected: %v, got: %v", http.MethodPut, r.Method)
		}
		h := r.Header
		if got := h.Get("x-ms-version"); got != APIVersion {
			t.Errorf("unexpected x-ms-version header. expected: %v, got: %v", APIVersion, got)
		}
		if got := h.Get("Date"); got != now.UTC().Format(time.RFC3339) {
			t.Errorf("unexpected Date header. expected: %v, got: %v", now.UTC().Format(time.RFC3339), got)
		}
		if got := h.Get("User-Agent"); got != DefaultUserAgent {
			t.Errorf("unexpected User-Agent. expected: %v, got: %v", DefaultUserAgent, got)
		}
		q := r.URL.Query()
		if got := q.Get("comp"); got != "blocklist" {
			t.Errorf("unexpected comp params. expected: blocklist, got: %v", got)
		}
		if r.ContentLength != int64(len(expected)) {
			t.Errorf("unexpected ContentLength. expected: %v, got: %v", len(expected), r.ContentLength)
		}
		got, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		if bytes.Compare(got, expected) != 0 {
			t.Error("post invalid xml")
		}
		w.WriteHeader(http.StatusCreated)
	})
	s := httptest.NewServer(m)
	defer s.Close()

	client, err := NewSASClient(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.PutBlockList(context.TODO(), blockList); err != nil {
		t.Error(err)
	}
}
