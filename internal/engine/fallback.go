package engine

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
)

//go:embed fallback.html
var fallbackHTMLContent string

var fallbackTemplate = template.Must(template.New("fallback").Parse(fallbackHTMLContent))

type ErrorDetails struct {
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
	Title      string `json:"title"`
	Message    string `json:"message"`
	Host       string `json:"host"`
	TunnelID   string `json:"tunnelId,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

func (e ErrorDetails) IsQuotaExceeded() bool {
	return e.StatusCode == http.StatusTooManyRequests ||
		e.StatusCode == 509 ||
		e.Error == "bandwidth_limit_exceeded" ||
		e.Error == "quota_exceeded"
}

func (e ErrorDetails) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound || e.Error == "tunnel_not_found"
}

func writeProxyError(response http.ResponseWriter, request *http.Request, details ErrorDetails) {
	accept := request.Header.Get("Accept")
	if strings.Contains(accept, "text/html") {
		response.Header().Set("Content-Type", "text/html; charset=utf-8")
		response.WriteHeader(details.StatusCode)
		_ = fallbackTemplate.Execute(response, details)
		return
	}

	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(details.StatusCode)
	_ = json.NewEncoder(response).Encode(details)
}
