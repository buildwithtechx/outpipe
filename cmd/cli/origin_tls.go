package main

import (
	"fmt"
	"net/http"

	"outpipe.dev/outpipe/pkg/client"
)

func prepareOriginClient(protocolName string, config client.OriginTLSConfig) (*http.Client, error) {
	if protocolName != "https" && (config.CAFile != "" || config.ServerName != "") {
		return nil, fmt.Errorf("origin TLS options require --protocol https")
	}
	origin, err := client.NewOriginHTTPClient(config)
	if err != nil {
		return nil, fmt.Errorf("configure local origin TLS: %w", err)
	}
	return origin, nil
}
