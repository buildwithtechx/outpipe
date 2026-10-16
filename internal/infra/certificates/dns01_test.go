package certificates

import (
	"crypto/ecdsa"
	"os"
	"path/filepath"
	"testing"
)

func TestNewDNS01IssuerRequiresConfiguration(t *testing.T) {
	if _, err := NewDNS01Issuer(ACMEConfig{CacheDir: t.TempDir()}); err == nil {
		t.Fatal("dns01 issuer accepted an empty contact email")
	}
	if _, err := NewDNS01Issuer(ACMEConfig{Email: "owner@example.com"}); err == nil {
		t.Fatal("dns01 issuer accepted an empty cache directory")
	}
	issuer, err := NewDNS01Issuer(ACMEConfig{Email: "owner@example.com", CacheDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	first, err := issuer.accountKey()
	if err != nil {
		t.Fatal(err)
	}
	second, err := issuer.accountKey()
	if err != nil {
		t.Fatal(err)
	}
	if !first.(*ecdsa.PrivateKey).Equal(second) {
		t.Fatal("account key was not stable across loads")
	}
	entries, err := os.ReadDir(filepath.Join(issuer.config.CacheDir))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "acme_account.key" {
		t.Fatalf("account key was not persisted: %v", entries)
	}
}
