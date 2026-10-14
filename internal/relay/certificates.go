package relay

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"outpipe.dev/outpipe/internal/infra/httpclient"
)

type DomainCertificate struct {
	Hostname    string
	ChainPEM    string
	KeyPEM      string
	ExpiresAt   time.Time
	WildcardFor string
}

type InternalCertificateFetcher struct {
	baseURL  string
	secret   string
	client   *http.Client
	mu       sync.Mutex
	certs    map[string]cachedDomainCertificate
	requests singleflight.Group
}

type cachedDomainCertificate struct {
	cert    *tls.Certificate
	expires time.Time
}

func NewInternalCertificateFetcher(baseURL, secret string, client *http.Client) (*InternalCertificateFetcher, error) {
	if strings.TrimSpace(baseURL) == "" || strings.TrimSpace(secret) == "" {
		return nil, fmt.Errorf("internal api url and secret are required")
	}
	if client == nil {
		client = httpclient.New(0)
	}
	return &InternalCertificateFetcher{baseURL: strings.TrimRight(baseURL, "/"), secret: secret, client: client, certs: make(map[string]cachedDomainCertificate)}, nil
}

func (f *InternalCertificateFetcher) CertificateForHostname(hostname string) (*tls.Certificate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return f.CertificateFor(ctx, hostname)
}

func (f *InternalCertificateFetcher) CertificateFor(ctx context.Context, hostname string) (*tls.Certificate, error) {
	host := strings.ToLower(strings.TrimSpace(strings.TrimSuffix(hostname, ".")))
	if host == "" {
		return nil, fmt.Errorf("hostname is required")
	}
	f.mu.Lock()
	cached, ok := f.certs[host]
	f.mu.Unlock()
	if ok && time.Now().Before(cached.expires) && cached.cert != nil {
		return cached.cert, nil
	}
	value, err, _ := f.requests.Do(host, func() (any, error) {
		lookupCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		certificate, err := f.fetch(lookupCtx, host)
		if err != nil {
			return nil, err
		}
		if len(f.certs) >= 4096 {
			now := time.Now()
			f.mu.Lock()
			for key, item := range f.certs {
				if now.After(item.expires) {
					delete(f.certs, key)
				}
			}
			f.mu.Unlock()
		}
		f.mu.Lock()
		if len(f.certs) < 4096 {
			f.certs[host] = cachedDomainCertificate{cert: certificate, expires: time.Now().Add(time.Minute)}
		}
		f.mu.Unlock()
		return certificate, nil
	})
	if err != nil {
		return nil, fmt.Errorf("load domain certificate: %w", err)
	}
	return value.(*tls.Certificate), nil
}

func (f *InternalCertificateFetcher) fetch(ctx context.Context, hostname string) (*tls.Certificate, error) {
	endpoint := f.baseURL + "/internal/domains/certificate?hostname=" + url.QueryEscape(hostname)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create domain certificate request: %w", err)
	}
	request.Header.Set("X-Internal-Secret", f.secret)
	response, err := f.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch domain certificate with api: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("domain certificate returned status %d", response.StatusCode)
	}
	var body struct {
		Hostname    string `json:"hostname"`
		Certificate string `json:"certificate"`
		Key         string `json:"key"`
		ExpiresAt   string `json:"expiresAt"`
		WildcardFor string `json:"wildcardFor"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode domain certificate response: %w", err)
	}
	if body.Certificate == "" || body.Key == "" {
		return nil, fmt.Errorf("domain certificate response is incomplete")
	}
	certificate, err := tls.X509KeyPair([]byte(body.Certificate), []byte(body.Key))
	if err != nil {
		return nil, fmt.Errorf("parse domain certificate: %w", err)
	}
	return &certificate, nil
}
