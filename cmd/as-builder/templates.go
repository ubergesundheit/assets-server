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
	"github.com/lpar/gzipped"
	"github.com/rakyll/statik/fs"
)

func main() {
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	logger := log.New(os.Stdout, "%[3]s: ", log.LstdFlags)
	logger.Printf("Starting %[3]s")

	server := &http.Server{
		Addr:         ":%[2]d",
		Handler:      wrapper(logger)(gzipped.FileServer(statikFS)),
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
		logger.Println("%[3]s is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown %[3]s: %%v\n", err)
		}
		close(done)
	}()

	logger.Println("%[3]s is listening at :%[2]d")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on :%[2]d: %%v\n", err)
	}

	<-done
	logger.Println("%[3]s stopped")
}

func wrapper(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if %[4]t {
				defer func() {
					remoteAddr := r.Header.Get("X-Forwarded-For")
					if remoteAddr == "" {
						remoteAddr = r.RemoteAddr
					}
					logger.Println(r.Method, r.URL.Path, remoteAddr, r.UserAgent())
				}()
				if r.URL.Path == "/" {
					r.URL.Path = "/index.html"
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
`
