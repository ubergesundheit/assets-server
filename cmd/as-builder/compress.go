package main

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/brotli/go/cbrotli"
)

func compressFiles(srcPath string) error {
	return filepath.Walk(srcPath, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
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
		if compress == false {
			return nil
		}

		// read file into buffer
		originalFileBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		// write gzip
		err = writeGzipFile(path, originalFileBytes)
		if err != nil {
			return err
		}

		// write brotli
		return writeBrotliFile(path, originalFileBytes)
	})
}

func writeBrotliFile(originalPath string, originalFileBytes []byte) error {
	var brotliBuffer bytes.Buffer
	brWriter := cbrotli.NewWriter(&brotliBuffer, cbrotli.WriterOptions{Quality: 11, LGWin: 0})
	// write bytes to brotliwriter to be brotlified
	_, err := brWriter.Write(originalFileBytes)
	if err != nil {
		return err
	}
	// close..
	if err = brWriter.Close(); err != nil {
		return err
	}
	// write to .gz file
	f, err := os.Create(originalPath + ".br")
	if err != nil {
		return err
	}
	defer f.Close()

	return ioutil.WriteFile(f.Name(), brotliBuffer.Bytes(), 0644)
}

func writeGzipFile(originalPath string, originalFileBytes []byte) error {
	var gzipBuffer bytes.Buffer
	gzipWriter, _ := gzip.NewWriterLevel(&gzipBuffer, 9)

	// write bytes to gzipwriter to be gzipped
	_, err := gzipWriter.Write(originalFileBytes)
	if err != nil {
		return err
	}
	// close..
	if err = gzipWriter.Close(); err != nil {
		return err
	}
	// write to .gz file
	f, err := os.Create(originalPath + ".gz")
	if err != nil {
		return err
	}
	defer f.Close()

	return ioutil.WriteFile(f.Name(), gzipBuffer.Bytes(), 0644)
}
