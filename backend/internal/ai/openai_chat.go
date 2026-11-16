package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultOpenAIChatURL = "https://api.openai.com/v1/chat/completions"

type OpenAIChatProvider struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

type OpenAIChatOption func(*OpenAIChatProvider)

func WithChatBaseURL(url string) OpenAIChatOption {
	return func(p *OpenAIChatProvider) {
		p.baseURL = strings.TrimRight(url, "/")
	}
}

func WithChatModel(model string) OpenAIChatOption {
	return func(p *OpenAIChatProvider) {
		p.model = model
	}
}

func WithChatHTTPClient(client *http.Client) OpenAIChatOption {
	return func(p *OpenAIChatProvider) {
		if client != nil {
			p.httpClient = client
		}
	}
}

func NewOpenAIChatProvider(
	apiKey string,
	options ...OpenAIChatOption,
) (*OpenAIChatProvider, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, ErrEmptyChatAPIKey
	}

	provider := &OpenAIChatProvider{
		apiKey:  strings.TrimSpace(apiKey),
		model:   "gpt-4o-mini",
		baseURL: defaultOpenAIChatURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}

	for _, option := range options {
		option(provider)
	}

	if strings.TrimSpace(provider.model) == "" {
		return nil, fmt.Errorf("chat model is required")
	}

	if strings.TrimSpace(provider.baseURL) == "" {
		return nil, fmt.Errorf("chat API URL is required")
	}

	return provider, nil
}

type openAIChatRequest struct {
	Model          string               `json:"model"`
	Messages       []ChatMessage        `json:"messages"`
	Temperature    float64              `json:"temperature"`
	ResponseFormat openAIResponseFormat `json:"response_format"`
}

type openAIResponseFormat struct {
	Type string `json:"type"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

type openAIChatError struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (p *OpenAIChatProvider) Complete(
	ctx context.Context,
	request ChatCompletionRequest,
) (string, error) {
	if len(request.Messages) == 0 {
		return "", ErrEmptyMessages
	}

	payload := openAIChatRequest{
		Model:          p.model,
		Messages:       request.Messages,
		Temperature:    request.Temperature,
		ResponseFormat: openAIResponseFormat{Type: "json_object"},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		p.baseURL,
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("create chat request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	response, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send chat request: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 5<<20))
	if err != nil {
		return "", fmt.Errorf("read chat response: %w", err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var apiError openAIChatError
		if json.Unmarshal(responseBody, &apiError) == nil &&
			apiError.Error.Message != "" {
			return "", fmt.Errorf(
				"OpenAI chat API returned HTTP %d: %s",
				response.StatusCode,
				apiError.Error.Message,
			)
		}

		return "", fmt.Errorf(
			"OpenAI chat API returned HTTP %d",
			response.StatusCode,
		)
	}

	var result openAIChatResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return "", fmt.Errorf("decode chat response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", ErrEmptyCompletion
	}

	content := strings.TrimSpace(result.Choices[0].Message.Content)
	if content == "" {
		return "", ErrEmptyCompletion
	}

	return content, nil
}
