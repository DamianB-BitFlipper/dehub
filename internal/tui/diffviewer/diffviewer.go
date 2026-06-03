package diffviewer

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cli/go-gh/v2/pkg/browser"
)

const (
	BuiltInPager        = "builtin:pipediffshub"
	heartbeatTimeout    = 30 * time.Second
	closeGrace          = 10 * time.Second
	heartbeatCheckEvery = 500 * time.Millisecond
)

//go:embed dist/* dist/assets/*
var embeddedDist embed.FS

type Options struct {
	Diff        []byte
	RefreshDiff func(context.Context) ([]byte, error)
	SourceURL   string
	Title       string
	BaseRefName string
	HeadRefName string
}

func IsBuiltInPager(pager string) bool {
	return strings.TrimSpace(pager) == BuiltInPager
}

func Open(ctx context.Context, opts Options) error {
	if len(strings.TrimSpace(string(opts.Diff))) == 0 {
		return errors.New("pipediffshub: no diff received")
	}

	dist, err := fs.Sub(embeddedDist, "dist")
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}

	serverCtx, cancelServerCtx := context.WithCancel(ctx)
	defer cancelServerCtx()

	state := &serverState{
		diff:        opts.Diff,
		refreshDiff: opts.RefreshDiff,
		dist:        dist,
		sourceURL:   opts.SourceURL,
		title:       opts.Title,
		baseRefName: opts.BaseRefName,
		headRefName: opts.HeadRefName,
		shutdown:    make(chan struct{}),
	}
	server := &http.Server{
		Handler: state,
		BaseContext: func(net.Listener) context.Context {
			return serverCtx
		},
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(listener)
	}()

	go state.watchHeartbeat()

	b := browser.New("", os.Stdout, os.Stdin)
	browserURL := fmt.Sprintf("http://%s/", listener.Addr().String())
	if err := b.Browse(browserURL); err != nil {
		state.requestShutdown()
		_ = server.Shutdown(context.Background())
		return err
	}

	select {
	case <-ctx.Done():
		state.requestShutdown()
	case <-state.shutdown:
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}

	err = <-serveErr
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

type serverState struct {
	diff        []byte
	refreshDiff func(context.Context) ([]byte, error)
	dist        fs.FS
	sourceURL   string
	title       string
	baseRefName string
	headRefName string

	mu                sync.Mutex
	receivedHeartbeat bool
	lastHeartbeatAt   time.Time
	closeTimer        *time.Timer
	shutdown          chan struct{}
	shutdownOnce      sync.Once
}

func (s *serverState) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/diff", "/api/diff":
		s.serveDiff(w, r)
	case "/meta":
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"sourceURL":   s.sourceURL,
			"title":       s.title,
			"baseRefName": s.baseRefName,
			"headRefName": s.headRefName,
		})
	case "/heartbeat":
		s.recordHeartbeat()
		w.WriteHeader(http.StatusNoContent)
	case "/close":
		s.scheduleClose()
		w.WriteHeader(http.StatusNoContent)
	default:
		s.serveStatic(w, r)
	}
}

func (s *serverState) serveDiff(w http.ResponseWriter, r *http.Request) {
	diff := s.diff
	if s.refreshDiff != nil {
		refreshedDiff, err := s.refreshDiff(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		diff = refreshedDiff
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write(diff)
}

func (s *serverState) recordHeartbeat() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.receivedHeartbeat = true
	s.lastHeartbeatAt = time.Now()
	if s.closeTimer != nil {
		s.closeTimer.Stop()
		s.closeTimer = nil
	}
}

func (s *serverState) scheduleClose() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closeTimer != nil {
		return
	}
	s.closeTimer = time.AfterFunc(closeGrace, s.requestShutdown)
}

func (s *serverState) watchHeartbeat() {
	ticker := time.NewTicker(heartbeatCheckEvery)
	defer ticker.Stop()

	for {
		select {
		case <-s.shutdown:
			return
		case <-ticker.C:
			if s.shouldStopForHeartbeat() {
				s.requestShutdown()
				return
			}
		}
	}
}

func (s *serverState) shouldStopForHeartbeat() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.receivedHeartbeat && time.Since(s.lastHeartbeatAt) >= heartbeatTimeout
}

func (s *serverState) requestShutdown() {
	s.shutdownOnce.Do(func() {
		close(s.shutdown)
	})
}

func (s *serverState) serveStatic(w http.ResponseWriter, r *http.Request) {
	filePath := cleanAssetPath(r.URL.Path)
	content, err := fs.ReadFile(s.dist, filePath)
	if err != nil {
		content, err = fs.ReadFile(s.dist, "index.html")
		filePath = "index.html"
	}
	if err != nil {
		http.NotFound(w, r)
		return
	}

	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = http.DetectContentType(content)
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", contentType)
	_, _ = w.Write(content)
}

func cleanAssetPath(rawPath string) string {
	if rawPath == "/" || rawPath == "" {
		return "index.html"
	}

	cleaned := path.Clean("/" + rawPath)
	return strings.TrimPrefix(cleaned, "/")
}
