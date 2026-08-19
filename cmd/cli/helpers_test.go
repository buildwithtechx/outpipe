package main

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	reader, writer, err := os.Pipe()

	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = writer

	var mu sync.Mutex
	output := make([]byte, 0, 4096)
	done := make(chan struct{})
	go func() {
		defer close(done)
		buffer := make([]byte, 4096)

		for {
			n, readErr := reader.Read(buffer)

			if n > 0 {
				mu.Lock()
				output = append(output, buffer[:n]...)
				mu.Unlock()
			}

			if readErr != nil {
				return
			}
		}
	}()

	fn()
	_ = writer.Close()
	os.Stdout = original
	<-done
	mu.Lock()
	value := string(output)
	mu.Unlock()

	return value
}

func cliTestEnv(t *testing.T, apiURL string) {
	t.Helper()
	t.Setenv("OUTPIPE_API_URL", apiURL)
	t.Setenv("OUTPIPE_RELAY_URL", "ws://relay.test")
	t.Setenv("OUTPIPE_CONFIG_PATH", filepath.Join(t.TempDir(), "config.json"))
}
