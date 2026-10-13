package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"outpipe.dev/outpipe/internal/engine"
)

const (
	captureSpoolMaxFiles = 4096
	captureSpoolMaxAge   = 72 * time.Hour
)

type captureSpool struct {
	dir      string
	maxFiles int64
	maxAge   time.Duration
	count    atomic.Int64
	now      func() time.Time
	delay    func(int) time.Duration
}

func newCaptureSpool(dir string) (*captureSpool, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, fmt.Errorf("capture spool directory is required")
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("create capture spool directory: %w", err)
	}
	spool := &captureSpool{dir: dir, maxFiles: captureSpoolMaxFiles, maxAge: captureSpoolMaxAge, now: time.Now}
	for _, file := range spool.pending() {
		if strings.HasSuffix(file, ".sending") {
			if err := os.Rename(filepath.Join(dir, file), strings.TrimSuffix(filepath.Join(dir, file), ".sending")); err != nil {
				return nil, fmt.Errorf("reclaim capture spool file: %w", err)
			}
		}
		spool.count.Add(1)
	}
	return spool, nil
}

func (r *captureRecorder) setSpoolDir(dir string) error {
	spool, err := newCaptureSpool(dir)
	if err != nil {
		return err
	}
	r.spool = spool
	return nil
}

func sanitizeSpoolID(id string) string {
	if id == "" {
		return uuid.NewString()
	}
	return strings.Map(func(c rune) rune {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' {
			return c
		}
		return '_'
	}, id)
}

func (s *captureSpool) store(capture engine.RequestCapture) (string, error) {
	if s.count.Load() >= s.maxFiles {
		return "", fmt.Errorf("capture spool is full")
	}
	name := fmt.Sprintf("%d-00-%s.json", s.now().UnixNano(), sanitizeSpoolID(capture.ID))
	body, err := json.Marshal(capture)
	if err != nil {
		return "", fmt.Errorf("encode spooled capture: %w", err)
	}
	tmp, err := os.CreateTemp(s.dir, "*.tmp")
	if err != nil {
		return "", fmt.Errorf("write capture spool file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(body); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return "", fmt.Errorf("write capture spool file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return "", fmt.Errorf("sync capture spool file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return "", fmt.Errorf("close capture spool file: %w", err)
	}
	path := filepath.Join(s.dir, name)
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return "", fmt.Errorf("commit capture spool file: %w", err)
	}
	s.count.Add(1)
	return path, nil
}

func (s *captureSpool) pending() []string {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".json.sending") {
			files = append(files, name)
		}
	}
	sort.Strings(files)
	return files
}

func spoolAttempt(name string) int {
	parts := strings.Split(strings.TrimSuffix(name, ".json"), "-")
	if len(parts) < 3 {
		return 0
	}
	attempt, err := strconv.Atoi(parts[1])
	if err != nil || attempt < 0 {
		return 0
	}
	return attempt
}

func spoolAttemptOfPath(path string) int {
	return spoolAttempt(strings.TrimSuffix(filepath.Base(path), ".sending"))
}

func spoolRetryDelay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	if attempt > 5 {
		attempt = 5
	}
	return time.Duration(1<<uint(attempt)) * time.Minute
}

func (s *captureSpool) retryDelay(attempt int) time.Duration {
	if s.delay != nil {
		return s.delay(attempt)
	}
	return spoolRetryDelay(attempt)
}

func (s *captureSpool) claim(name string) (string, error) {
	return s.claimPath(filepath.Join(s.dir, name))
}

func (s *captureSpool) claimPath(path string) (string, error) {
	claimed := path + ".sending"
	if err := os.Rename(path, claimed); err != nil {
		return "", fmt.Errorf("claim capture spool file: %w", err)
	}
	return claimed, nil
}

func (s *captureSpool) release(claimed string, attempts int, availableAt time.Time) {
	name := strings.TrimSuffix(filepath.Base(claimed), ".sending")
	parts := strings.Split(name, "-")
	created := fmt.Sprint(s.now().UnixNano())
	id := strings.TrimSuffix(name, ".json")
	if len(parts) >= 3 {
		created = parts[0]
		id = strings.Join(parts[2:], "-")
	}
	next := fmt.Sprintf("%s-%02d-%s.json", created, attempts, id)
	if err := os.Rename(claimed, filepath.Join(s.dir, next)); err != nil {
		return
	}
	_ = os.Chtimes(filepath.Join(s.dir, next), s.now(), availableAt)
}

func (s *captureSpool) complete(claimed string) {
	_ = os.Remove(claimed)
	s.count.Add(-1)
}

func (s *captureSpool) evict(now time.Time) int {
	evicted := 0
	for _, name := range s.pending() {
		if strings.HasSuffix(name, ".sending") {
			continue
		}
		path := filepath.Join(s.dir, name)
		aged := false
		if parts := strings.Split(strings.TrimSuffix(name, ".json"), "-"); len(parts) >= 1 {
			if ts, err := strconv.ParseInt(parts[0], 10, 64); err == nil && now.Sub(time.Unix(0, ts)) > s.maxAge {
				aged = true
			}
		}
		if !aged {
			if info, err := os.Stat(path); err == nil && now.Sub(info.ModTime()) > s.maxAge {
				aged = true
			}
		}
		if aged {
			_ = os.Remove(path)
			s.count.Add(-1)
			evicted++
		}
	}
	return evicted
}
