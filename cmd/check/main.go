package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"outpipe.dev/outpipe/internal/config"
)

var version = "dev"

func main() {
	cfg, err := config.LoadCheck()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.Service.InternalAPIURL+"/readyz", nil)
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("X-Internal-Secret", cfg.Service.InternalAPISecret)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		log.Fatal(fmt.Errorf("api readiness check returned status %d", response.StatusCode))
	}
	fmt.Println("ready")
}
