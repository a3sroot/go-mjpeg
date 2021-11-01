//go:build ignore
// +build ignore

package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/mattn/go-mjpeg"
)

var (
	addr     = flag.String("addr", ":8080", "Server address")
	path     = flag.String("path","","Save path")
	interval = flag.Duration("interval", 5*time.Minute, "interval")
)

func main() {
	flag.Parse()
	if *path == "" {
		flag.Usage()
		os.Exit(1)
	}

	stream := mjpeg.NewStream(*path,*interval)
	for  {
		stream.Update()
	}
	var wg sync.WaitGroup
	wg.Add(1)

	http.HandleFunc("/jpeg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(stream.Current())
	})

	http.HandleFunc("/mjpeg", stream.ServeHTTP)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<img src="/mjpeg" />`))
	})

	server := &http.Server{Addr: *addr}
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	go func() {
		<-sc
		server.Shutdown(context.Background())
	}()
	log.Println("open server")
	server.ListenAndServe()
	stream.Close()

	wg.Wait()
}
