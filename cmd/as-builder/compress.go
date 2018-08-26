package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/brotli/go/cbrotli"
)

func compressFiles(srcPath string) (string, error) {
	etags := strings.Builder{}

	err := filepath.Walk(srcPath, func(path string, fi os.FileInfo, err error) error {
		// Ignore directories
		if fi.IsDir() {
			return nil
		}
		compress := false
		// check if the file has an extension which we want to compress
		for _, ext := range extensionsToCompress {
			if strings.HasSuffix(fi.Name(), ext) {
				compress = true
				break
			}
		}

		relativePath := strings.TrimPrefix(path, srcPath)

		// read file into buffer
		originalFileBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		if compress {
			// write brotli and compute etag
			brotliEtag, err := writeBrotliFile(path, relativePath, originalFileBytes)
			if err != nil {
				return err
			}
			_, err = etags.WriteString(brotliEtag)
			if err != nil {
				return err
			}
			// write gzip and compute etag
			gzipEtag, err := writeGzipFile(path, relativePath, originalFileBytes)
			if err != nil {
				return err
			}
			_, err = etags.WriteString(gzipEtag)
			if err != nil {
				return err
			}
		}

		// compute plain etag
		_, err = etags.WriteString(computeEtag(originalFileBytes, relativePath))
		if err != nil {
			return err
		}

		return nil
	})

	return etags.String(), err
}

func computeEtag(fileContents []byte, relativePath string) string {
	return fmt.Sprintf("\"%s\": \"\\\"%x\\\"\",\n", relativePath, sha256.Sum256(fileContents))
}

func writeBrotliFile(originalPath, relativePath string, originalFileBytes []byte) (string, error) {
	var brotliBuffer bytes.Buffer
	brWriter := cbrotli.NewWriter(&brotliBuffer, cbrotli.WriterOptions{Quality: 11, LGWin: 0})
	// write bytes to brotliwriter to be brotlified
	_, err := brWriter.Write(originalFileBytes)
	if err != nil {
		return "", err
	}
	// close..
	if err = brWriter.Close(); err != nil {
		return "", err
	}
	// write to .gz file
	f, err := os.Create(originalPath + ".br")
	if err != nil {
		return "", err
	}
	defer f.Close()

	brotliBytes := brotliBuffer.Bytes()
	return computeEtag(brotliBytes, relativePath+".br"), ioutil.WriteFile(f.Name(), brotliBytes, 0644)
}

func writeGzipFile(originalPath, relativePath string, originalFileBytes []byte) (string, error) {
	var gzipBuffer bytes.Buffer
	gzipWriter, _ := gzip.NewWriterLevel(&gzipBuffer, 9)

	// write bytes to gzipwriter to be gzipped
	_, err := gzipWriter.Write(originalFileBytes)
	if err != nil {
		return "", err
	}
	// close..
	if err = gzipWriter.Close(); err != nil {
		return "", err
	}
	// write to .gz file
	f, err := os.Create(originalPath + ".gz")
	if err != nil {
		return "", err
	}
	defer f.Close()

	gzipBytes := gzipBuffer.Bytes()
	return computeEtag(gzipBytes, relativePath+".gz"), ioutil.WriteFile(f.Name(), gzipBytes, 0644)
}
