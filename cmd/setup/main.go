package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type downloadInfo struct {
	URL  string
	Name string
}

func main() {
	if err := _main(); err != nil {
		log.Fatal(err)
	}
}

func _main() error {
	var testdataDir string
	flag.StringVar(&testdataDir, "d", "testdata", "testdata directory")

	var filename string
	flag.StringVar(&filename, "f", "filelist.json", "download list path")

	flag.Parse()

	f, err := os.Open(filename)
	if err != nil {
		return errors.Wrapf(err, "download list open failed")
	}
	defer f.Close()

	var downloadInfoList []downloadInfo
	if err := json.NewDecoder(f).Decode(&downloadInfoList); err != nil {
		return errors.Wrap(err, "donwload list decode failed")
	}

	for _, info := range downloadInfoList {
		func() {
			log.Printf("[INFO] donwloading %v ...", info.URL)
			resp, err := http.Get(info.URL)
			if err != nil {
				log.Printf("[INFO] donwload failed, url: %v, reason: %v", info.URL, err)
				return
			}
			defer resp.Body.Close()

			fpath := filepath.Join(testdataDir, info.Name)
			log.Printf("[INFO] save to %v ...", fpath)

			f, err := os.Create(fpath)
			if err != nil {
				log.Printf("[INFO] file create failed, path: %v, reason: %v", fpath, err)
			}
			defer f.Close()

			io.Copy(f, resp.Body)
		}()
	}

	log.Print("[INFO] complete")
	return nil
}
