package main

const mainGoTemplate = `package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "./statik"
	"github.com/rakyll/statik/fs"
)

var logger *log.Logger

func main() {
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	logger = log.New(os.Stdout, "%[2]s: ", log.LstdFlags)
	logger.Printf("Starting %[2]s")

	server := &http.Server{
		Addr:         ":%[1]d",
		Handler:      FileServer(statikFS),
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		logger.Println("%[2]s is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown %[2]s: %%v\n", err)
		}
		close(done)
	}()

	logger.Println("%[2]s is listening at :%[1]d")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on :%[1]d: %%v\n", err)
	}

	<-done
	logger.Println("%[2]s stopped")
}
`

const serverGoTemplate = `package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/golang/gddo/httputil/header"
)

const (
	gzipEncoding  = "gzip"
	gzipExtension = ".gz"

	brotliEncoding  = "br"
	brotliExtension = ".br"
)

type fileHandler struct {
	root http.FileSystem
}

var etags map[string]string

// FileServer is a drop-in replacement for Go's standard http.FileServer
// which adds support for static resources precompressed with gzip, at
// the cost of removing the support for directory browsing.
//
// If file filename.ext has a compressed version filename.ext.gz alongside
// it, if the client indicates that it accepts gzip-compressed data, and
// if the .gz file can be opened, then the compressed version of the file
// will be sent to the client. Otherwise the request is passed on to
// http.ServeContent, and the raw (uncompressed) version is used.
//
// It is up to you to ensure that the compressed and uncompressed versions
// of files match and have sensible timestamps.
//
// Compressed or not, requests are fulfilled using http.ServeContent, and
// details like accept ranges and content-type sniffing are handled by that
// method.
func FileServer(root http.FileSystem) http.Handler {
	etags = map[string]string{
		%[2]s
	}
	return &fileHandler{root}
}

func acceptable(r *http.Request, encoding string) bool {
	for _, aspec := range header.ParseAccept(r.Header, "Accept-Encoding") {
		if aspec.Value == encoding && aspec.Q == 0.0 {
			return false
		}
		if (aspec.Value == encoding || aspec.Value == "*") && aspec.Q > 0.0 {
			return true
		}
	}
	return false
}

func (f *fileHandler) openAndStat(path string) (http.File, os.FileInfo, error) {
	file, err := f.root.Open(path)
	var info os.FileInfo
	// This slightly weird variable reuse is so we can get 100%% test coverage
	// without having to come up with a test file that can be opened, yet
	// fails to stat.
	if err == nil {
		info, err = file.Stat()
	}
	if err != nil {
		return file, nil, err
	}
	if info.IsDir() {
		return file, nil, fmt.Errorf("%%s is directory", path)
	}
	return file, info, nil
}

%[4]s

func (f *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// logging enabled?
	if %[1]t {
		defer func() {
			remoteAddr := r.Header.Get("X-Forwarded-For")
			if remoteAddr == "" {
				remoteAddr = r.RemoteAddr
			}
			logger.Println(r.Method, r.URL.Path, remoteAddr, r.UserAgent())
		}()
	}

	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}
	fpath := path.Clean(upath)

	// rewriting
	if %[3]t {
		fpath = findRewrite(fpath)
	}
	if strings.HasSuffix(fpath, "/") {
		upath = upath + "index.html"
		r.URL.Path = upath
		fpath = fpath + "index.html"
	} else if strings.HasSuffix(upath, "/") {
		upath = upath + "index.html"
		r.URL.Path = upath
		fpath = fpath + "/index.html"
	}
	// Try for a compressed version if appropriate
	var file http.File
	var err error
	var info os.FileInfo
	var fPathLoaded string

	foundAcceptable := false

	if acceptable(r, brotliEncoding) {
		fPathLoaded = fpath + brotliExtension
		file, info, err = f.openAndStat(fPathLoaded)
		if err == nil {
			foundAcceptable = true
			w.Header().Set("Content-Encoding", brotliEncoding)
		}
	}

	if !foundAcceptable && acceptable(r, gzipEncoding) {
		fPathLoaded = fpath + gzipExtension
		file, info, err = f.openAndStat(fPathLoaded)
		if err == nil {
			foundAcceptable = true
			w.Header().Set("Content-Encoding", gzipEncoding)
		}
	}
	// If we didn't manage to open a compressed version, try for uncompressed
	if !foundAcceptable {
		file, info, err = f.openAndStat(fpath)
		fPathLoaded = fpath
	}
	if err != nil {
		// Doesn't exist compressed or uncompressed
		// custom 404 page or default handler..
		%[5]s
		return
	}
	defer file.Close()

	if strings.HasSuffix(fpath, ".html") {
		w.Header().Set("Cache-Control", "public, max-age=0, must-revalidate")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=31536000,immutable")
		if etag, ok := etags[fPathLoaded]; ok {
			w.Header().Set("Etag", etag)
		}
	}
	http.ServeContent(w, r, fpath, info.ModTime(), file)
}
`
