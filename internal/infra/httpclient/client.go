package httpclient

import (
	"net/http"
	"time"
)

const DefaultTimeout = 15 * time.Second

func New(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	return &http.Client{Timeout: timeout}
}
