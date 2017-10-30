package blob

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/orisano/httpc"
	"github.com/pkg/errors"
)

const (
	APIVersion = "2017-04-17"
)

var DefaultUserAgent = fmt.Sprintf("Go/%s (%s-%s) go-ams.blob", runtime.Version(), runtime.GOARCH, runtime.GOOS)

type SASClient struct {
	rb         *httpc.RequestBuilder
	httpClient *http.Client

	userAgent string
	logger    *log.Logger
	debug     bool

	TimeNow func() time.Time
}

func NewSASClient(rawurl string, opts ...clientOption) (*SASClient, error) {
	options := &clientOptions{
		Client:    http.DefaultClient,
		UserAgent: DefaultUserAgent,
		Logger:    log.New(ioutil.Discard, "", log.Lshortfile),
		Debug:     false,
	}
	for _, opt := range opts {
		opt(options)
	}

	h := make(http.Header)
	h.Set("x-ms-version", APIVersion)
	h.Set("User-Agent", options.UserAgent)

	rb, err := httpc.NewRequestBuilder(rawurl, h)
	if err != nil {
		return nil, err
	}

	if options.Debug {
		httpc.InjectDebugTransport(options.Client, os.Stderr)
	}

	return &SASClient{
		rb:         rb,
		httpClient: options.Client,
		userAgent:  options.UserAgent,
		logger:     options.Logger,
		debug:      options.Debug,
		TimeNow:    time.Now,
	}, nil
}

func withDate(t time.Time) httpc.RequestOption {
	return func(o *httpc.RequestOptions) error {
		o.Header.Set("Date", t.UTC().Format(time.RFC3339))
		return nil
	}
}

func (c *SASClient) PutBlob(ctx context.Context, blob io.Reader, blockID string) error {
	if ctx == nil {
		return errors.New("missing ctx")
	}
	if blob == nil {
		return errors.New("missing blob")
	}
	if len(blockID) == 0 {
		return errors.New("missing blockID")
	}
	req, err := c.rb.NewRequest(ctx, http.MethodPut, "",
		withDate(c.TimeNow()),
		httpc.SetHeaderField("x-ms-blob-type", "BlockBlob"),
		httpc.AddQuery("comp", "block"),
		httpc.AddQuery("blockid", blockID),
		httpc.WithBinary(blob),
		httpc.EnforceContentLength,
	)
	if err != nil {
		return errors.Wrap(err, "failed to construct http request")
	}
	c.logger.Print("[INFO] put blob ...")
	resp, err := httpc.Retry(c.httpClient, req)
	if err != nil {
		return errors.Wrap(err, "failed to http request")
	}
	defer resp.Body.Close()

	if got := resp.StatusCode; got != http.StatusCreated {
		return errors.Errorf("unexpected status code. expected: %v, but got: %v", http.StatusCreated, got)
	}

	c.logger.Print("[INFO] completed")
	return nil
}

func (c *SASClient) PutBlockList(ctx context.Context, blockList []string) error {
	if ctx == nil {
		return errors.New("missing ctx")
	}
	if len(blockList) == 0 {
		return errors.New("missing blockList")
	}
	req, err := c.rb.NewRequest(ctx, http.MethodPut, "",
		withDate(c.TimeNow()),
		httpc.AddQuery("comp", "blocklist"),
		httpc.WithXML(&BlockList{Blocks: blockList}),
		httpc.EnforceContentLength,
	)
	if err != nil {
		return errors.Wrap(err, "failed to construct http request")
	}
	c.logger.Print("[INFO] put block list ...")
	resp, err := httpc.Retry(c.httpClient, req)
	if err != nil {
		return errors.Wrap(err, "failed to http request")
	}
	defer resp.Body.Close()

	if got := resp.StatusCode; got != http.StatusCreated {
		return errors.Errorf("unexpected status code. expected: %v, but got: %v", http.StatusCreated, got)
	}

	c.logger.Print("[INFO] completed")
	return nil
}

type UploadError struct {
	errs []error
	m    *sync.Mutex
}

func newUploadError() *UploadError {
	return &UploadError{
		m: new(sync.Mutex),
	}
}

func (e *UploadError) add(err error) {
	if err != nil {
		e.m.Lock()
		e.errs = append(e.errs, err)
		e.m.Unlock()
	}
}

func (e *UploadError) Error() string {
	return fmt.Sprintf("failed to upload: (%d error occurred)", len(e.errs))
}

func (e *UploadError) Errors() []error {
	return e.errs
}

func (c *SASClient) Upload(ctx context.Context, r io.Reader, chunkSize int64, workers uint) (int64, error) {
	if ctx == nil {
		return 0, errors.New("missing ctx")
	}
	if r == nil {
		return 0, errors.New("missing r")
	}
	if chunkSize <= 0 {
		return 0, errors.New("chunkSize must be greater than 0")
	}
	if workers == 0 {
		return 0, errors.New("workers must be greater then 0")
	}

	type job struct {
		blockID string
		r       io.Reader
	}

	jobs := make(chan job, workers)
	uploadErr := newUploadError()
	wg := new(sync.WaitGroup)
	wg.Add(int(workers))
	for w := uint(0); w < workers; w++ {
		go func() {
			for job := range jobs {
				err := c.PutBlob(ctx, job.r, job.blockID)
				if err != nil {
					uploadErr.add(err)
				}
			}
			wg.Done()
		}()
	}

	var blockList []string
	contentLength := int64(0)
	bi := 0
	for {
		bi++
		b := bytes.NewBuffer(make([]byte, 0, chunkSize))
		n, err := io.CopyN(b, r, chunkSize)
		if n == 0 {
			break
		}
		if err != nil {
			close(jobs)
			wg.Done()
			return 0, errors.Wrap(err, "failed to read r")
		}

		contentLength += n
		blockID := buildBlockID(bi)
		blockList = append(blockList, blockID)
		jobs <- job{blockID, b}

		if n < chunkSize {
			break
		}
	}
	close(jobs)
	wg.Done()

	if uploadErr.Errors() != nil {
		return 0, uploadErr
	}

	if err := c.PutBlockList(ctx, blockList); err != nil {
		return 0, errors.Wrap(err, "failed to put blocklist")
	}

	return contentLength, nil
}
