package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/gianghp123/SonaVoice/api/internal/core/errors"
)

type IHttpClient interface {
	Do(
		ctx context.Context,
		method string,
		url string,
		headers map[string]string,
		body any,
	) ([]byte, int, *errors.AppError)
}

type httpClient struct {
	client *http.Client
}

func NewHttpClient() IHttpClient {
	return &httpClient{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (h *httpClient) Do(
	ctx context.Context,
	method string,
	url string,
	headers map[string]string,
	body any,
) ([]byte, int, *errors.AppError) {
	var reader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			sentry.CaptureException(err)
			return nil, 0, errors.Internal("failed to marshal request body")
		}

		reader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		url,
		reader,
	)
	if err != nil {
		sentry.CaptureException(err)
		return nil, 0, errors.Internal("failed to create request")
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		sentry.CaptureException(err)
		return nil, 0, errors.Internal("failed to execute request")
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		sentry.CaptureException(err)
		return nil, 0, errors.Internal("failed to read response")
	}

	return responseBody, resp.StatusCode, nil
}
