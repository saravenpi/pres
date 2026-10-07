package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/saravenpi/pres/internal/parser"
	"github.com/saravenpi/pres/internal/renderer"
)

type Options struct {
	Dir  string
	Port int
}

func Serve(opts Options) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	p, err := parser.Parse(opts.Dir)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", opts.Dir, err)
	}

	mu := &sync.Mutex{}
	var html []byte

	rebuild := func(pp *parser.Presentation) error {
		b, err := renderer.Render(pp, false, nil)
		if err != nil {
			return err
		}
		mu.Lock()
		html = b
		mu.Unlock()
		return nil
	}

	if err := rebuild(p); err != nil {
		return fmt.Errorf("rendering: %w", err)
	}

	reloadClients := map[chan struct{}]bool{}
	reloadMu := &sync.Mutex{}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		mu.Lock()
		data := make([]byte, len(html))
		copy(data, html)
		mu.Unlock()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		w.Write(data)
	})

	mux.HandleFunc("/__pres_reload", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		flusher, ok := w.(http.Flusher)
		if !ok {
			return
		}
		ch := make(chan struct{}, 1)
		reloadMu.Lock()
		reloadClients[ch] = true
		reloadMu.Unlock()
		defer func() {
			reloadMu.Lock()
			delete(reloadClients, ch)
			reloadMu.Unlock()
		}()
		for {
			select {
			case <-r.Context().Done():
				return
			case _, ok := <-ch:
				if !ok {
					return
				}
			}
			fmt.Fprintf(w, "event: reload\ndata: reload\n\n")
			flusher.Flush()
		}
	})

	go watch(ctx, opts.Dir, &p, mu, rebuild, reloadMu, reloadClients)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", opts.Port))
	if err != nil {
		return fmt.Errorf("listen :%d: %w", opts.Port, err)
	}
	addr := ln.Addr().String()
	fmt.Printf("▸ Serving on http://%s\n", addr)
	fmt.Printf("✓ Watching for changes...\n")

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	srv := &http.Server{
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	err = srv.Serve(ln)
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func watch(
	ctx context.Context,
	dir string,
	p **parser.Presentation,
	mu *sync.Mutex,
	rebuild func(*parser.Presentation) error,
	reloadMu *sync.Mutex,
	reloadClients map[chan struct{}]bool,
) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var lastMod time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		mod, err := dirModTime(dir)
		if err != nil {
			continue
		}
		if !mod.After(lastMod) || lastMod.IsZero() {
			lastMod = mod
			continue
		}
		lastMod = mod

		newP, err := parser.Parse(dir)
		if err != nil {
			log.Printf("parse error: %v", err)
			continue
		}

		mu.Lock()
		cur := newP
		*p = cur
		mu.Unlock()

		if err := rebuild(cur); err != nil {
			log.Printf("render error: %v", err)
			continue
		}

		reloadMu.Lock()
		for ch := range reloadClients {
			select {
			case ch <- struct{}{}:
			default:
			}
		}
		reloadMu.Unlock()
	}
}

func dirModTime(dir string) (time.Time, error) {
	var latest time.Time
	entries, err := os.ReadDir(dir)
	if err != nil {
		return latest, fmt.Errorf("reading %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(latest) {
			latest = info.ModTime()
		}
	}
	return latest, nil
}