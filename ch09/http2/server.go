package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/senicko/network-programming-with-go/ch09/handlers"
	"github.com/senicko/network-programming-with-go/ch09/middleware"
)

var (
	addr  = flag.String("listen", "127.0.0.1:8080", "listen address")
	cert  = flag.String("cert", "", "certificate")
	pkey  = flag.String("key", "", "private key")
	files = flag.String("files", "./files", "static files directory")
)

func main() {
	flag.Parse()

	fmt.Printf("listen: %s\n", *addr)
	fmt.Printf("cert: %s\n", *cert)
	fmt.Printf("key: %s\n", *pkey)
	fmt.Printf("files: %s\n", *files)

	err := run(*addr, *files, *cert, *pkey)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server gracefully shutdown")
}

func run(addr, files, cert, pkey string) error {
	mux := http.NewServeMux()
	mux.Handle("/static/",
		http.StripPrefix("/static/",
			middleware.RestrictPrefix(".",
				http.FileServer(
					http.Dir(files),
				),
			),
		),
	)

	mux.Handle("/", handlers.Methods{
		http.MethodGet: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if pusher, ok := w.(http.Pusher); ok {
				targets := []string{
					"/static/styles.css",
					"/static/shield.svg",
				}

				for _, target := range targets {
					if err := pusher.Push(target, nil); err != nil {
						log.Printf("%s push failed: %v", target, err)
					}
				}
			}

			http.ServeFile(w, r, filepath.Join(files, "index.html"))
		}),
	})

	mux.Handle("/2", handlers.Methods{
		http.MethodGet: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(files, "index.html"))
		}),
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: 30 * time.Second,
	}

	done := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		for {
			if <-c == os.Interrupt {
				if err := srv.Shutdown(context.Background()); err != nil {
					log.Printf("shutdown error: %v", err)
				}
				close(done)
				return
			}
		}
	}()

	log.Printf("Serving files in %q over %s\n", files, srv.Addr)

	var err error
	if cert != "" && pkey != "" {
		log.Println("TLS enabled")
		err = srv.ListenAndServeTLS(cert, pkey)
	} else {
		err = srv.ListenAndServe()
	}

	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	if err == http.ErrServerClosed {
		err = nil
	}

	<-done

	return nil
}
