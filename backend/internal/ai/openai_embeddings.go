package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"
)

const defaultOpenAIEmbeddingURL = "https://api.openai.com/v1/embeddings"

type OpenAIEmbeddingProvider struct {
	apiKey     string
	model      string
	baseURL    string
	dimensions int
	httpClient *http.Client
}

type OpenAIEmbeddingOption func(*OpenAIEmbeddingProvider)

func WithEmbeddingBaseURL(url string) OpenAIEmbeddingOption {
	return func(p *OpenAIEmbeddingProvider) {
		p.baseURL = strings.TrimRight(url, "/")
	}
}

func WithEmbeddingDimensions(dimensions int) OpenAIEmbeddingOption {
	return func(p *OpenAIEmbeddingProvider) {
		p.dimensions = dimensions
	}
}

func WithEmbeddingHTTPClient(client *http.Client) OpenAIEmbeddingOption {
	return func(p *OpenAIEmbeddingProvider) {
		if client != nil {
			p.httpClient = client
		}
	}
}

func NewOpenAIEmbeddingProvider(
	apiKey string,
	options ...OpenAIEmbeddingOption,
) (*OpenAIEmbeddingProvider, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, ErrEmptyAPIKey
	}

	provider := &OpenAIEmbeddingProvider{
		apiKey:  strings.TrimSpace(apiKey),
		model:   "text-embedding-3-small",
		baseURL: defaultOpenAIEmbeddingURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, option := range options {
		option(provider)
	}

	if provider.dimensions < 0 {
		return nil, fmt.Errorf("embedding dimensions cannot be negative")
	}

	if strings.TrimSpace(provider.baseURL) == "" {
		return nil, fmt.Errorf("embedding API URL is required")
	}

	return provider, nil
}

type openAIEmbeddingRequest struct {
	Model      string   `json:"model"`
	Input      []string `json:"input"`
	Dimensions int      `json:"dimensions,omitempty"`
}

type openAIEmbeddingResponse struct {
	Data []openAIEmbeddingData `json:"data"`
}

type openAIEmbeddingData struct {
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

type openAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func (p *OpenAIEmbeddingProvider) Embed(
	ctx context.Context,
	texts []string,
) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	for i, text := range texts {
		if strings.TrimSpace(text) == "" {
			return nil, fmt.Errorf("%w: item %d", ErrEmptyInput, i)
		}
	}

	payload := openAIEmbeddingRequest{
		Model:      p.model,
		Input:      texts,
		Dimensions: p.dimensions,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal embedding request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		p.baseURL,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("create embedding request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	response, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send embedding request: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 20<<20))
	if err != nil {
		return nil, fmt.Errorf("read embedding response: %w", err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var apiError openAIErrorResponse
		if json.Unmarshal(responseBody, &apiError) == nil &&
			apiError.Error.Message != "" {
			return nil, fmt.Errorf(
				"%w: HTTP %d: %s",
				ErrUnexpectedStatus,
				response.StatusCode,
				apiError.Error.Message,
			)
		}

		return nil, fmt.Errorf(
			"%w: HTTP %d",
			ErrUnexpectedStatus,
			response.StatusCode,
		)
	}

	var result openAIEmbeddingResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	if len(result.Data) != len(texts) {
		return nil, fmt.Errorf(
			"%w: expected %d vectors, received %d",
			ErrInvalidResponse,
			len(texts),
			len(result.Data),
		)
	}

	// The API response should be mapped by index rather than relying
	// on the order of returned items.
	sort.Slice(result.Data, func(i, j int) bool {
		return result.Data[i].Index < result.Data[j].Index
	})

	vectors := make([][]float32, len(texts))
	expectedDimensions := p.dimensions

	for i, item := range result.Data {
		if item.Index != i {
			return nil, fmt.Errorf(
				"%w: missing or duplicate index %d",
				ErrInvalidResponse,
				i,
			)
		}

		if len(item.Embedding) == 0 {
			return nil, fmt.Errorf(
				"%w: vector %d is empty",
				ErrInvalidEmbedding,
				i,
			)
		}

		if expectedDimensions == 0 {
			expectedDimensions = len(item.Embedding)
		}

		if len(item.Embedding) != expectedDimensions {
			return nil, fmt.Errorf(
				"%w: vector %d has %d dimensions; expected %d",
				ErrInvalidEmbedding,
				i,
				len(item.Embedding),
				expectedDimensions,
			)
		}

		for _, value := range item.Embedding {
			if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
				return nil, fmt.Errorf(
					"%w: vector %d contains a non-finite value",
					ErrInvalidEmbedding,
					i,
				)
			}
		}

		vectors[i] = item.Embedding
	}

	return vectors, nil
}
