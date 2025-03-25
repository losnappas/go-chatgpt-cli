package conversation

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type Role string

func (r Role) String() string {
	return string(r)
}

const (
	RoleUser      Role = "User"
	RoleAssistant Role = "Assistant"
	RoleSystem    Role = "System"
)

func NewRole(s string) (Role, error) {
	r := Role(s)
	switch r {
	case RoleUser, RoleAssistant, RoleSystem:
		return r, nil
	default:
		return "", fmt.Errorf("invalid role: %q", s)
	}
}

type ConversationTurn struct {
	Content string
	Role    Role
}

func (c ConversationTurn) String() string {
	return fmt.Sprintf(`# %v

%v

`, c.Role, c.Content)
}

type Conversation struct {
	Messages []ConversationTurn
}

func (c *Conversation) Empty() bool {
	return len(c.Messages) == 0
}

func (c *Conversation) Append(role Role, msg string) *ConversationTurn {
	m := strings.TrimSpace(msg)
	if m != "" {
		t := ConversationTurn{Content: m, Role: role}
		c.Messages = append(c.Messages, t)
		return &t
	}
	return nil
}

func (c *Conversation) AppendUserMessage(msg string) *ConversationTurn {
	return c.Append(RoleUser, msg)
}

func (c *Conversation) AppendSystemMessage(msg string) *ConversationTurn {
	return c.Append(RoleSystem, msg)
}

type Parser interface {
	Parse(context.Context, io.Reader) (*Conversation, error)
}
