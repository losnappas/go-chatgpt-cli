package llm

import (
	"context"
	"fmt"

	"github.com/losnappas/go-chatgpt-cli/internal/conversation"
	"google.golang.org/genai"
)

type GoogleClient struct {
	ClientOptions
}

func (c *GoogleClient) Respond(
	ctx context.Context,
	convo *conversation.Conversation,
) <-chan string {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)

	client, _ := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  c.ApiKey,
		Backend: genai.BackendGeminiAPI,
	})

	system, chat := c.destructureConversation(convo)

	stream := client.Models.GenerateContentStream(
		ctx,
		c.Model,
		chat,
		&genai.GenerateContentConfig{
			SystemInstruction: system,
		},
	)

	outChan := make(chan string)

	go func() {
		defer close(outChan)
		defer cancel()

		for chunk, err := range stream {
			if err != nil {
				outChan <- "Error: "
				outChan <- err.Error()
				continue
			}
			outChan <- chunk.Text()
		}
	}()

	return outChan
}

func (c *GoogleClient) destructureConversation(
	convo *conversation.Conversation,
) (*genai.Content, []*genai.Content) {
	out := make([]*genai.Content, 0, len(convo.Messages))
	var system *genai.Content

	for _, msg := range convo.Messages {
		if msg.Role == conversation.RoleSystem {
			system = genai.Text(msg.Content)[0]
			continue
		}

		out = append(out, &genai.Content{
			Role: role(msg.Role),
			Parts: []*genai.Part{
				{
					Text: msg.Content,
				},
			},
		})
	}
	return system, out
}

func role(r conversation.Role) string {
	switch r {
	case conversation.RoleAssistant:
		return "model"
	case conversation.RoleUser:
		return "user"
	}
	panic(fmt.Sprintf("unexpected role: %s", r))
}
