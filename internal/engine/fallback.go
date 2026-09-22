package engine

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"
)

type ErrorDetails struct {
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
	Title      string `json:"title"`
	Message    string `json:"message"`
	Host       string `json:"host"`
	TunnelID   string `json:"tunnelId,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

func writeProxyError(response http.ResponseWriter, request *http.Request, details ErrorDetails) {
	accept := request.Header.Get("Accept")
	if strings.Contains(accept, "text/html") {
		response.Header().Set("Content-Type", "text/html; charset=utf-8")
		response.WriteHeader(details.StatusCode)
		_, _ = response.Write([]byte(renderFallbackHTML(details)))
		return
	}

	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(details.StatusCode)
	_ = json.NewEncoder(response).Encode(details)
}

func renderFallbackHTML(details ErrorDetails) string {
	safeTitle := html.EscapeString(details.Title)
	safeMessage := html.EscapeString(details.Message)
	safeHost := html.EscapeString(details.Host)
	safeReason := html.EscapeString(details.Reason)

	reasonBlock := ""
	if safeReason != "" {
		reasonBlock = fmt.Sprintf(`<div class="reason"><code>%s</code></div>`, safeReason)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%d %s — Outpipe</title>
  <style>
    :root {
      --bg: #090a0f;
      --card: #12141c;
      --border: #1e2230;
      --text: #f3f4f6;
      --muted: #9ca3af;
      --accent: #6366f1;
      --danger: #ef4444;
      --code-bg: #0d0f17;
    }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
      background: var(--bg);
      color: var(--text);
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 1.5rem;
    }
    .container {
      max-width: 580px;
      width: 100%%;
      background: var(--card);
      border: 1px solid var(--border);
      border-radius: 16px;
      padding: 2.5rem;
      box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
    }
    .badge {
      display: inline-flex;
      align-items: center;
      gap: 0.5rem;
      padding: 0.25rem 0.75rem;
      border-radius: 9999px;
      font-size: 0.75rem;
      font-weight: 600;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      background: rgba(239, 68, 68, 0.15);
      color: var(--danger);
      border: 1px solid rgba(239, 68, 68, 0.3);
      margin-bottom: 1.25rem;
    }
    .badge-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%%;
      background: var(--danger);
    }
    h1 {
      font-size: 1.5rem;
      font-weight: 700;
      color: var(--text);
      margin-bottom: 0.75rem;
      line-height: 1.3;
    }
    p {
      color: var(--muted);
      font-size: 0.95rem;
      line-height: 1.6;
      margin-bottom: 1.25rem;
    }
    .meta {
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
      background: var(--code-bg);
      border: 1px solid var(--border);
      border-radius: 10px;
      padding: 1rem;
      margin-bottom: 1.5rem;
      font-size: 0.85rem;
    }
    .meta-row {
      display: flex;
      justify-content: space-between;
      color: var(--muted);
    }
    .meta-val {
      color: var(--text);
      font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    }
    .reason {
      background: rgba(239, 68, 68, 0.08);
      border-left: 3px solid var(--danger);
      padding: 0.75rem 1rem;
      margin-bottom: 1.5rem;
      border-radius: 0 8px 8px 0;
      font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, monospace;
      font-size: 0.825rem;
      color: #fca5a5;
      word-break: break-all;
    }
    .tips {
      background: rgba(99, 102, 241, 0.06);
      border: 1px solid rgba(99, 102, 241, 0.2);
      border-radius: 10px;
      padding: 1rem 1.25rem;
      margin-bottom: 1.5rem;
    }
    .tips h3 {
      font-size: 0.85rem;
      font-weight: 600;
      color: var(--accent);
      text-transform: uppercase;
      letter-spacing: 0.05em;
      margin-bottom: 0.5rem;
    }
    .tips ul {
      list-style-position: inside;
      color: var(--muted);
      font-size: 0.875rem;
      line-height: 1.6;
    }
    .actions {
      display: flex;
      gap: 0.75rem;
    }
    .btn {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      padding: 0.65rem 1.25rem;
      border-radius: 8px;
      font-size: 0.875rem;
      font-weight: 500;
      text-decoration: none;
      cursor: pointer;
      transition: all 0.15s ease;
      border: none;
    }
    .btn-primary {
      background: var(--accent);
      color: white;
    }
    .btn-primary:hover {
      background: #4f46e5;
    }
    .btn-secondary {
      background: var(--border);
      color: var(--text);
    }
    .btn-secondary:hover {
      background: #282e42;
    }
    .brand {
      margin-top: 2rem;
      text-align: center;
      font-size: 0.75rem;
      color: #6b7280;
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="badge">
      <span class="badge-dot"></span>
      HTTP %d • %s
    </div>
    <h1>%s</h1>
    <p>%s</p>

    %s

    <div class="meta">
      <div class="meta-row"><span>Tunnel Host</span><span class="meta-val">%s</span></div>
      <div class="meta-row"><span>Status</span><span class="meta-val">Local Endpoint Offline</span></div>
    </div>

    <div class="tips">
      <h3>Troubleshooting</h3>
      <ul>
        <li>Ensure your local development server is running and listening on the configured port.</li>
        <li>Check your terminal or Outpipe desktop client to ensure the agent is connected.</li>
        <li>Confirm your local service is bound to <code>127.0.0.1</code> or <code>0.0.0.0</code>.</li>
      </ul>
    </div>

    <div class="actions">
      <button class="btn btn-primary" onclick="window.location.reload()">Retry Connection</button>
      <a class="btn btn-secondary" href="https://outpipe.app" target="_blank" rel="noopener">Outpipe Home</a>
    </div>

    <div class="brand">
      Powered by Outpipe Tunnel Relay
    </div>
  </div>
</body>
</html>`, details.StatusCode, safeTitle, details.StatusCode, safeTitle, safeTitle, safeMessage, reasonBlock, safeHost)
}
