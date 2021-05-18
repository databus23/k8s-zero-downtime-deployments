package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func serve(ctx context.Context) (err error) {

	startupCompleted := time.Now().Add(startupDelay)

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log.Println(r.Method, r.URL.Path, "?", r.URL.RawQuery)
			if time.Now().Before(startupCompleted) {
				w.WriteHeader(500)
				fmt.Fprintf(w, "starting up...")
				return
			}

			i, err := strconv.Atoi(r.URL.Query().Get("wait"))
			if err == nil {
				time.Sleep(time.Duration(i) * time.Second)
			}
			fmt.Fprintf(w, "okay")
		},
	))

	srv := &http.Server{
		Addr:    ":6969",
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen:%+s\n", err)
		}
	}()

	log.Printf("server started %s", srv.Addr)
	<-ctx.Done()

	log.Printf("server stopped")

	if gracefulShutdownTimeout > 0 {
		ctxShutDown, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		defer cancel()

		if err = srv.Shutdown(ctxShutDown); err != nil {
			log.Fatalf("Graceful server shutdown Failed:%s", err)
		}

	} else {
		if err := srv.Close(); err != nil {
			log.Fatalf("Stopping server failed: %s", err)
		}
	}

	log.Printf("server exited properly")

	if err == http.ErrServerClosed {
		err = nil
	}

	return
}

var startupDelay time.Duration
var gracefulShutdownTimeout time.Duration

func main() {
	flag.DurationVar(&startupDelay, "startup-delay", 5*time.Second, "Artifically delay the startup of the webserver")
	flag.DurationVar(&gracefulShutdownTimeout, "graceful-shutdown-timeout", 0, "Timeout for graceful shutdown. Default: no graceful shutdown")
	flag.Parse()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		oscall := <-c
		log.Printf("system call:%+v", oscall)
		cancel()
	}()

	if err := serve(ctx); err != nil {
		log.Printf("failed to serve:+%v\n", err)
	}
}
