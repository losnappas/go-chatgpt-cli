package conversation_test

import (
	"context"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/losnappas/go-chatgpt-cli/internal/conversation"
)

func TestMarkdownParser_Parse(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		input   io.Reader
		want    *conversation.Conversation
		wantErr bool
	}{
		{
			name: "Basic",
			input: strings.NewReader(`---
model: etst
---

# User

# your text here

stuff

# More text here

stuff2

# Assistant

works

# User

# User

# User

gg

# Assistant

asd

# Assistant

ggz`),
			want: &conversation.Conversation{
				Messages: []conversation.ConversationTurn{
					{
						Content: `# your text here

stuff

# More text here

stuff2`,
						Role: "User",
					},
					{
						Content: `works`,
						Role:    "Assistant",
					},
					{
						Content: `gg`,
						Role:    "User",
					},
					{
						Content: `asd


ggz`,
						Role: "Assistant",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m conversation.MarkdownParser
			got, gotErr := m.Parse(context.Background(), tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				if !tt.wantErr {
					t.Errorf("Parse() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Parse() succeeded unexpectedly")
			}
		})
	}
}
