package history

import (
	"fmt"
	"io"

	"github.com/losnappas/go-chatgpt-cli/internal/conversation"
)

type HistoryWriter interface {
	Write(*conversation.ConversationTurn)
	WriteString(string)
	WriteHeading(conversation.Role)
	Close()
}

type MarkdownHistory struct {
	OutputHandle io.Writer
}

func (mh *MarkdownHistory) Write(m *conversation.ConversationTurn) {
	mh.OutputHandle.Write([]byte(m.String()))
}

func (mh *MarkdownHistory) WriteHeading(t conversation.Role) {
	mh.OutputHandle.Write(fmt.Appendf(nil, "# %v\n\n", t))
}

func (mh *MarkdownHistory) WriteString(s string) {
	mh.OutputHandle.Write([]byte(s))
}

func (mh *MarkdownHistory) Close() {
	mh.OutputHandle.Write([]byte("\n\n"))
}
