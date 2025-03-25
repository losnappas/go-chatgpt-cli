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

	messages := m.messages(text)

	return &Conversation{
		Messages: messages,
	}, nil
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
