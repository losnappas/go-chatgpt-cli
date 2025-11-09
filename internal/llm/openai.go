package llm

import (
	"context"

	"github.com/losnappas/go-chatgpt-cli/internal/conversation"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OpenaiClient struct {
	ClientOptions
	BaseURL string
}

func (c *OpenaiClient) Respond(
	ctx context.Context,
	convo *conversation.Conversation,
) <-chan string {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)

	opts := []option.RequestOption{
		option.WithAPIKey(c.ApiKey),
	}
	if c.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(c.BaseURL))
	}

	client := openai.NewClient(opts...)

	stream := client.Chat.Completions.NewStreaming(
		ctx,
		openai.ChatCompletionNewParams{
			Messages:        c.destructureConversation(convo),
			Model:           openai.ChatModel(c.Model),
			ReasoningEffort: "low",
		},
	)

	outChan := make(chan string)

	go func() {
		defer close(outChan)
		defer cancel()
		for stream.Next() {
			chunk := stream.Current()

			if len(chunk.Choices) > 0 {
				outChan <- chunk.Choices[0].Delta.Content
			}
		}
		if stream.Err() != nil {
			outChan <- "Error: "
			outChan <- stream.Err().Error()
		}
	}()

	return outChan
}

func (c *OpenaiClient) destructureConversation(
	convo *conversation.Conversation,
) []openai.ChatCompletionMessageParamUnion {
	out := make([]openai.ChatCompletionMessageParamUnion, 0, len(convo.Messages))

	for _, msg := range convo.Messages {
		switch msg.Role {
		case conversation.RoleAssistant:
			out = append(out, openai.AssistantMessage(msg.Content))

		case conversation.RoleUser:
			out = append(out, openai.UserMessage(msg.Content))

		case conversation.RoleSystem:
			out = append(out, openai.SystemMessage(msg.Content))

		default:
			panic("unexpected openai conversation role")
		}
	}

	return out
}
