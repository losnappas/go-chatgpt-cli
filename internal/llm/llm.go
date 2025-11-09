package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/losnappas/go-chatgpt-cli/internal/conversation"
)

type LlmClient interface {
	Respond(context.Context, *conversation.Conversation) <-chan string
}

type ClientOptions struct {
	ApiKey  string
	Model   string
	Timeout time.Duration
}

func NewLlmClient(apiKeys []string, providerModel string) (LlmClient, error) {
	modelProvider, model := parseSeparator(providerModel, "/")
	var key string
	for _, apiKey := range apiKeys {
		if after, ok := strings.CutPrefix(apiKey, modelProvider+"="); ok {
			key = after
			break
		}
	}
	if key == "" || model == "" {
		return nil, errors.New("missing api key or model")
	}

	defaultOptions := ClientOptions{
		ApiKey:  key,
		Model:   model,
		Timeout: time.Second * 180,
	}

	switch modelProvider {
	case "openai":
		return &OpenaiClient{
			ClientOptions: defaultOptions,
		}, nil

	case "deepseek":
		return &OpenaiClient{
			ClientOptions: defaultOptions,
			BaseURL:       "https://api.deepseek.com",
		}, nil

	case "google":
		return &GoogleClient{
			ClientOptions: defaultOptions,
		}, nil
	}
	return nil, fmt.Errorf("unexpected provider: %v\n", modelProvider)
}

func parseSeparator(keystring, sep string) (string, string) {
	key := strings.SplitN(keystring, sep, 2)
	if len(key) != 2 {
		return "", ""
	}
	provider := key[0]
	apiKey := key[1]
	return provider, apiKey
}
