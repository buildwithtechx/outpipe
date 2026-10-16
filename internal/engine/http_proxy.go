package engine

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"outpipe.dev/outpipe/pkg/protocol"
)

type HTTPProxy struct {
	baseDomain   string
	router       *RequestRouter
	maxBodyBytes int64
	recorder     UsageRecorder
	capture      CaptureRecorder
}

type UsageMeasurement struct {
	OrganizationID string
	TunnelID       string
	EventType      string
	Bytes          int64
	Connections    int
	Method         string
	Path           string
	StatusCode     int
	DurationMillis int64
	ResponseBytes  int64
	ClientIP       *string `json:"clientIp,omitempty"`
}

type UsageRecorder interface {
	Record(context.Context, UsageMeasurement) error
}

func NewHTTPProxy(baseDomain string, router *RequestRouter, maxBodyBytes int64) (*HTTPProxy, error) {
	if baseDomain == "" {
		return nil, fmt.Errorf("tunnel base domain is required")
	}
	if router == nil {
		return nil, fmt.Errorf("request router is required")
	}
	return &HTTPProxy{baseDomain: strings.ToLower(strings.TrimSuffix(baseDomain, ".")), router: router, maxBodyBytes: maxBodyBytes}, nil
}

func (p *HTTPProxy) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	started := time.Now()
	route, ok := resolveRouteKey(request.Host, p.baseDomain)
	if !ok {
		writeProxyError(response, request, ErrorDetails{
			StatusCode: http.StatusNotFound,
			Error:      "tunnel_not_found",
			Title:      "Tunnel Not Found",
			Message:    "The requested tunnel hostname is not registered on Outpipe or has been revoked.",
			Host:       request.Host,
		})
		return
	}

	passwordHash, protected := p.router.PasswordHash(route)
	protected = protected && passwordHash != ""

	if protected && !authorizedRequest(request, passwordHash) {
		response.Header().Set("WWW-Authenticate", `Basic realm="Outpipe"`)
		response.WriteHeader(http.StatusUnauthorized)
		p.recordRequest(request, route, http.StatusUnauthorized, 0, started)
		return
	}

	p.refreshCapture(request.Context(), route)
	body, err := readBody(request.Body, p.maxBodyBytes)
	if err != nil {
		if err == errBodyTooLarge {
			writeProxyError(response, request, ErrorDetails{
				StatusCode: http.StatusRequestEntityTooLarge,
				Error:      "body_too_large",
				Title:      "Request Body Too Large",
				Message:    "The request payload exceeded the maximum allowed size for this tunnel.",
				Host:       request.Host,
				TunnelID:   route,
			})
			p.recordRequest(request, route, http.StatusRequestEntityTooLarge, 0, started)
			return
		}

		writeProxyError(response, request, ErrorDetails{
			StatusCode: http.StatusBadRequest,
			Error:      "bad_request",
			Title:      "Bad Request",
			Message:    "Unable to read incoming request payload.",
			Host:       request.Host,
			TunnelID:   route,
		})
		p.recordRequest(request, route, http.StatusBadRequest, 0, started)
		return
	}

	forwarded, err := p.router.ForwardHTTP(request.Context(), route, protocol.HTTPRequest{
		Method:  request.Method,
		Path:    request.URL.RequestURI(),
		Headers: requestHeaders(request.Header),
		Body:    body,
	})
	if err != nil {
		writeProxyError(response, request, ErrorDetails{
			StatusCode: http.StatusBadGateway,
			Error:      "tunnel_unavailable",
			Title:      "Tunnel Offline",
			Message:    "Outpipe could not reach an active local agent for this tunnel. The agent may be disconnected or asleep.",
			Host:       request.Host,
			TunnelID:   route,
			Reason:     err.Error(),
		})
		p.recordRequest(request, route, http.StatusBadGateway, 0, started)
		return
	}

	if forwarded.Error != "" {
		writeProxyError(response, request, ErrorDetails{
			StatusCode: http.StatusBadGateway,
			Error:      "local_service_unreachable",
			Title:      "Local Server Unreachable",
			Message:    "Outpipe agent is connected, but your local application failed to respond or the local port is closed.",
			Host:       request.Host,
			TunnelID:   route,
			Reason:     forwarded.Error,
		})
		p.recordRequest(request, route, http.StatusBadGateway, 0, started)
		return
	}
	if err := p.captureRequest(request, route, body, forwarded, started); err != nil {
		p.router.SetCaptureEnabled(route, false)
	}

	writeResponseHeaders(response.Header(), forwarded.Headers)
	status := forwarded.StatusCode
	if status < http.StatusContinue || status > 599 {
		status = http.StatusBadGateway
	}

	response.WriteHeader(status)
	if forwarded.Body == "" {
		p.recordRequest(request, route, status, 0, started)
		return
	}

	decoded, err := base64.StdEncoding.DecodeString(forwarded.Body)
	if err != nil {
		p.recordRequest(request, route, http.StatusBadGateway, 0, started)
		return
	}

	_, _ = response.Write(decoded)
	p.recordRequest(request, route, status, int64(len(decoded)), started)
}

func (p *HTTPProxy) SetUsageRecorder(recorder UsageRecorder) {
	p.recorder = recorder
}

func (p *HTTPProxy) recordRequest(request *http.Request, tunnelID string, statusCode int, responseBytes int64, started time.Time) {
	if p.recorder == nil {
		return
	}
	organizationID, ok := p.router.OrganizationID(tunnelID)
	if !ok {
		return
	}
	clientAddress := clientIP(request)
	var clientAddressValue *string
	if clientAddress != "" {
		clientAddressValue = &clientAddress
	}
	_ = p.recorder.Record(request.Context(), UsageMeasurement{OrganizationID: organizationID, TunnelID: tunnelID, EventType: "request", Bytes: responseBytes, Method: request.Method, Path: request.URL.Path, StatusCode: statusCode, DurationMillis: time.Since(started).Milliseconds(), ResponseBytes: responseBytes, ClientIP: clientAddressValue})
}
