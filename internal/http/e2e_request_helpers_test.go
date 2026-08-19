package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func (s *e2eStack) request(t *testing.T, app *fiber.App, method, path string, headers map[string]string, body string) *http.Response {
	t.Helper()

	var reader io.Reader

	if body != "" {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Accept", "application/json")

	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	response, err := app.Test(req, -1)

	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}

	return response
}

func (s *e2eStack) sessionCookie(t *testing.T, response *http.Response) string {
	t.Helper()

	for _, cookie := range response.Cookies() {

		if cookie.Name == "outpipe_session" {
			return cookie.Value
		}
	}

	return ""
}
