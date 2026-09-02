package client

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/infra/httpclient"
	"outpipe.dev/outpipe/pkg/protocol"
)

func (c *RelayConnection) forwardHTTP(ctx context.Context, targetURL string, incoming protocol.HTTPRequest) protocol.HTTPResponse {
	body, err := base64.StdEncoding.DecodeString(incoming.Body)

	if err != nil {
		return protocol.HTTPResponse{Error: "invalid request body"}
	}

	request, err := http.NewRequestWithContext(ctx, incoming.Method, strings.TrimRight(targetURL, "/")+incoming.Path, strings.NewReader(string(body)))

	if err != nil {
		return protocol.HTTPResponse{Error: err.Error()}
	}

	for key, values := range incoming.Headers {

		for _, value := range values {
			request.Header.Add(key, value)
		}
	}

	response, err := httpclient.New(90 * time.Second).Do(request)

	if err != nil {
		return protocol.HTTPResponse{Error: err.Error()}
	}

	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, 16<<20))

	if err != nil {
		return protocol.HTTPResponse{Error: err.Error()}
	}

	headers := make(map[string][]string, len(response.Header))

	for key, values := range response.Header {
		headers[key] = append([]string(nil), values...)
	}

	return protocol.HTTPResponse{StatusCode: response.StatusCode, Headers: headers, Body: base64.StdEncoding.EncodeToString(data)}
}
