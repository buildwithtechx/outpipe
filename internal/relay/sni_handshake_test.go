package relay

import (
	"crypto/tls"
	"net"
	"os"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/infra/certificates"
)

type staticCertificateProvider struct {
	cert *tls.Certificate
}

func (s *staticCertificateProvider) CertificateForHostname(hostname string) (*tls.Certificate, error) {
	if hostname == "app.example.com" {
		return s.cert, nil
	}
	return nil, errNoStaticCertificate
}

type staticCertError string

func (e staticCertError) Error() string { return string(e) }

const errNoStaticCertificate = staticCertError("no certificate for hostname")

func TestSNIDispatcherServesCustomDomainTLSHandshake(t *testing.T) {
	customPEM, customKey := serveFetcherCertificate(t, "app.example.com")
	custom, err := tls.X509KeyPair([]byte(customPEM), []byte(customKey))
	if err != nil {
		t.Fatal(err)
	}
	fallbackPEM, fallbackKey := serveFetcherCertificate(t, "relay.example.com")
	dir := t.TempDir()
	writeFile := func(name, content string) string {
		path := dir + "/" + name
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	fallbackCertFile := writeFile("fallback.crt", fallbackPEM)
	fallbackKeyFile := writeFile("fallback.key", fallbackKey)
	fallback, err := certificates.NewRotatingCertificate(fallbackCertFile, fallbackKeyFile)
	if err != nil {
		t.Fatal(err)
	}
	dispatcher, err := certificates.NewSNIDispatcher(fallback, &staticCertificateProvider{cert: &custom})
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	server := tls.NewListener(listener, dispatcher.Config())
	defer server.Close()
	go func() {
		for {
			conn, err := server.Accept()
			if err != nil {
				return
			}
			tlsConn, ok := conn.(*tls.Conn)
			if !ok {
				conn.Close()
				continue
			}
			_ = tlsConn.SetDeadline(time.Now().Add(5 * time.Second))
			if err := tlsConn.Handshake(); err != nil {
				conn.Close()
				continue
			}
			conn.Close()
		}
	}()

	dial := func(serverName string) *tls.Certificate {
		t.Helper()
		conn, err := tls.Dial("tcp", listener.Addr().String(), &tls.Config{InsecureSkipVerify: true, ServerName: serverName})
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		state := conn.ConnectionState()
		if len(state.PeerCertificates) == 0 {
			t.Fatal("server presented no certificate")
		}
		return &tls.Certificate{Certificate: [][]byte{state.PeerCertificates[0].Raw}}
	}

	customServed := dial("app.example.com")
	if len(customServed.Certificate) == 0 || string(customServed.Certificate[0]) != string(custom.Certificate[0]) {
		t.Fatal("custom domain did not receive its LE chain over SNI")
	}
	fallbackServed := dial("unknown.example.com")
	fallbackLoaded, err := certificates.NewRotatingCertificate(fallbackCertFile, fallbackKeyFile)
	if err != nil {
		t.Fatal(err)
	}
	want, err := fallbackLoaded.GetCertificate(nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(fallbackServed.Certificate[0]) != string(want.Certificate[0]) {
		t.Fatal("unknown hostname did not receive the mounted fallback chain")
	}
}
