package ai

import (
	"context"
	"errors"
)

var (
	ErrEmptyChatAPIKey = errors.New("chat API key is required")
	ErrEmptyMessages   = errors.New("chat messages cannot be empty")
	ErrEmptyCompletion = errors.New("chat completion is empty")
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Messages    []ChatMessage
	Temperature float64
}

type LLMProvider interface {
	Complete(
		ctx context.Context,
		request ChatCompletionRequest,
	) (string, error)
}
