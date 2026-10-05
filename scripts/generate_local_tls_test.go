package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateLocalTLSPreservesValidPair(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := generateLocalTLS(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join("data", "tls", "localhost.key")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := generateLocalTLS(); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("existing key was replaced")
	}
}

func TestCertificateWriteFailureRemovesOnlyNewKey(t *testing.T) {
	dir := t.TempDir()
	certPath, keyPath := filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem")
	if err := os.WriteFile(certPath, []byte("existing certificate"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := saveLocalTLSPair(certPath, keyPath, []byte("cert"), []byte("key")); err == nil {
		t.Fatal("expected exclusive certificate write to fail")
	}
	if _, err := os.Stat(keyPath); !os.IsNotExist(err) {
		t.Fatalf("incomplete key was retained: %v", err)
	}
	data, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "existing certificate" {
		t.Fatal("existing certificate was modified")
	}
}
