package conversation

import (
	"context"
	"io"
	"regexp"
	"strings"
)

type MarkdownParser struct{}

func NewMarkdownParser() *MarkdownParser {
	return &MarkdownParser{}
}

func (m *MarkdownParser) Parse(ctx context.Context, input io.Reader) (*Conversation, error) {
	textB, err := io.ReadAll(input)
	if err != nil {
		return nil, err
	}
	text := string(textB)

	model, content := m.parseFrontmatter(text)
	messages := m.messages(content)

	return &Conversation{
		Model:    model,
		Messages: messages,
	}, nil
}

func (m *MarkdownParser) parseFrontmatter(text string) (string, string) {
	parts := strings.SplitN(text, "---\n", 3)
	if len(parts) < 3 {
		return "", text
	}

	var model string
	frontmatter := parts[1]
	for line := range strings.SplitSeq(frontmatter, "\n") {
		if after, ok := strings.CutPrefix(line, "model:"); ok {
			model = strings.TrimSpace(after)
		}
	}
	content := parts[2]

	return model, strings.TrimSpace(content)
}

var roleRegexp = regexp.MustCompile("^# (User|Assistant|System)$")

func (m *MarkdownParser) messages(data string) []ConversationTurn {
	lines := strings.Split(data, "\n")
	turns := []ConversationTurn{}
	currentTurn := &ConversationTurn{}

	for _, line := range lines {
		role := roleRegexp.FindStringSubmatch(line)
		if len(role) != 2 {
			currentTurn.Content += line + "\n"
			continue
		}

		_role, err := NewRole(role[1])
		if err != nil {
			continue
		}
		if _role == currentTurn.Role {
			continue
		}

		turns = appendTurn(turns, currentTurn)
		currentTurn = &ConversationTurn{Role: Role(_role)}
	}
	turns = appendTurn(turns, currentTurn)

	return turns
}

func appendTurn(turns []ConversationTurn, currentTurn *ConversationTurn) []ConversationTurn {
	currentTurn.Content = strings.TrimSpace(currentTurn.Content)
	if currentTurn.Content != "" && currentTurn.Role != "" {
		turns = append(turns, *currentTurn)
	}
	return turns
}
