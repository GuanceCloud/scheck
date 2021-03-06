// Package install :download func
package install

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type writeCounter struct {
	total   uint64
	current uint64
	last    float64
}

func (wc *writeCounter) Write(p []byte) (n int, err error) {
	n = len(p)
	wc.current += uint64(n)
	wc.last += float64(n)
	wc.PrintProgress()
	return n, err
}

func (wc *writeCounter) PrintProgress() {
	if wc.last > float64(wc.total)*0.01 || wc.current == wc.total { // update progress-bar each 1%
		wc.last = 0.0
	}
}

func Download(cli *http.Client, from, to string, isGzip, progress, downloadOnly bool) error {
	req, err := http.NewRequest("GET", from, nil)
	if err != nil {
		return err
	}
	if isGzip {
		req.Header.Add("Accept-Encoding", "gzip")
	}

	resp, err := cli.Do(req)
	if err != nil {
		l.Errorf("err %v", err)
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	progbar := &writeCounter{
		total: uint64(resp.ContentLength),
	}

	if downloadOnly {
		return doDownload(io.TeeReader(resp.Body, progbar), to)
	}
	if !progress {
		return doExtract(resp.Body, to)
	}
	return doExtract(io.TeeReader(resp.Body, progbar), to)
}

func doDownload(r io.Reader, to string) error {
	f, err := os.OpenFile(filepath.Clean(to), os.O_CREATE|os.O_RDWR, os.ModeAppend|os.ModePerm)
	if err != nil {
		l.Errorf("open file err=%v to=%s", err, to)
		return err
	}

	if _, err := io.Copy(f, r); err != nil { //nolint:gosec
		return err
	}

	return f.Close()
}

func ExtractDatakit(gz, to string) error {
	data, err := os.Open(filepath.Clean(gz))
	if err != nil {
		l.Fatalf("open file %s failed: %s", gz, err)
	}

	defer func() {
		_ = data.Close()
	}()

	return doExtract(data, to)
}

func doExtract(r io.Reader, to string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		l.Error(err)
		return err
	}

	defer func() {
		_ = gzr.Close()
	}()
	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			l.Error(err)
			return err
		}
		if hdr == nil {
			continue
		}

		target := filepath.Join(to, filepath.Clean(hdr.Name))

		switch hdr.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, os.ModeDir|os.ModePerm); err != nil {
					l.Error(err)
					return err
				}
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), os.ModeDir|os.ModePerm); err != nil {
				l.Error(err)
				return err
			}

			// TODO: lock file before extracting, to avoid `text file busy` error
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode)) // nolint
			if err != nil {
				l.Error(err)
				return err
			}

			// #nosec
			if _, err := io.Copy(f, tr); err != nil {
				l.Error(err)
				return err
			}

			if err := f.Close(); err != nil {
				l.Warnf("f.Close(): %v, ignored", err)
			}
		default:
			l.Warnf("unexpected file %s", target)
		}
	}
}
