package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"

	"outpipe.dev/outpipe/internal/infra/certificates"
	"outpipe.dev/outpipe/internal/relay"
)

func relayServingListener(address, certFile, keyFile, internalAPIURL, internalSecret string) (net.Listener, error) {
	fallback, err := certificates.NewRotatingCertificate(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	var provider certificates.CertificateProvider
	if internalAPIURL != "" && internalSecret != "" {
		fetcher, err := relay.NewInternalCertificateFetcher(internalAPIURL, internalSecret, nil)
		if err != nil {
			log.Printf("warning: domain certificate fetching is disabled: %v", err)
		} else {
			provider = fetcher
		}
	}
	dispatcher, err := certificates.NewSNIDispatcher(fallback, provider)
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("listen on %s: %w", address, err)
	}
	return tls.NewListener(listener, dispatcher.Config()), nil
}
