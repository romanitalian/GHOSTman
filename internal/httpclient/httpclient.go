package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	// Shared HTTP client with timeout
	Client = &http.Client{
		Timeout: 10 * time.Second,
	}
)

// SendRequest sends an HTTP request and returns status, body, and error
func SendRequest(rq *http.Request) (statusLine string, prettyBody string, isError bool, err error) {
	resp, err := Client.Do(rq)
	if err != nil {
		return "", "", true, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	const maxResponseSize = 2 * 1024 * 1024 // 2 MB
	limitedReader := io.LimitReader(resp.Body, maxResponseSize)
	bodyRS, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", "", true, fmt.Errorf("error reading response: %v", err)
	}

	statusLine = fmt.Sprintf("HTTP %d %s\n", resp.StatusCode, resp.Status)

	if resp.StatusCode >= 400 {
		return statusLine, string(bodyRS), true, nil
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, bodyRS, "", "    "); err != nil {
		return statusLine, string(bodyRS), false, nil
	}

	return statusLine, prettyJSON.String(), false, nil
}

// NewRequest creates a new http.Request from method, url, body, and headers string
func NewRequest(method, url, body, headers string) (*http.Request, error) {
	rq, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}
	hdrs := strings.Split(headers, "\n")
	for _, h := range hdrs {
		if h == "" {
			continue
		}
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			rq.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}
	return rq, nil
}
