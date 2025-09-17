package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/losnappas/go-chatgpt-cli/internal/conversation"
	"github.com/losnappas/go-chatgpt-cli/internal/history"
	"github.com/losnappas/go-chatgpt-cli/internal/llm"
	"github.com/losnappas/go-chatgpt-cli/internal/renderer"
)

var (
	model       = flag.String("model", "openai/gpt-5-chat-latest", "The provider/model to use")
	apiKey      = flag.String("api-key", "", "The API key to use, as provider=api_key")
	historyFile = flag.String("history-file", "", "Required. Read chat history from markdown file")
	clear       = flag.Bool("c", false, "Clear history")
	editor      = flag.Bool("editor", false, "Open history with $EDITOR")
	system      = flag.String("system-prompt", "", "The LLM system prompt. Overridden by history file")
)

func main() {
	ctx := context.Background()
	flag.Parse()
	err := run(ctx)
	if err != nil {
		log.Fatalln(err)
	}
}

func run(ctx context.Context) error {
	if *historyFile == "" {
		return errors.New("history-file is a required argument")
	}
	if *clear {
		os.Remove(*historyFile)
	}
	if *apiKey == "" {
		return errors.New("Missing api key")
	}
	if *editor {
		return runEditor(*historyFile)
	}

	// Get positional arguments
	args := flag.Args()

	input := getStdinData()

	parser := conversation.NewMarkdownParser()

	// Make history dir if not exist.
	dir := filepath.Dir(*historyFile)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("Failed to create directory: %v\n", err)
	}

	f, err := os.OpenFile(*historyFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	convo, err := parser.Parse(ctx, f)
	if err != nil {
		return err
	}
	historyWriter := &history.MarkdownHistory{
		OutputHandle: f,
	}
	rend, err := renderer.NewPrinter()
	if err != nil {
		return err
	}
	defer rend.Close()

	if convo.Empty() {
		if msg := convo.AppendSystemMessage(*system); msg != nil {
			historyWriter.Write(msg)
		}
	}

	userMessage := strings.Join(args, " ")
	userMessage += "\n\n" + input
	msg := convo.AppendUserMessage(userMessage)
	if msg != nil {
		historyWriter.Write(msg)
	}

	client, err := llm.NewLlmClient(*apiKey, *model)
	if err != nil {
		return err
	}

	historyWriter.WriteHeading(conversation.RoleAssistant)

	response := client.Respond(ctx, convo)

	for value := range response {
		rend.Print(value)
		historyWriter.WriteString(value)
	}
	historyWriter.Close()
	return nil
}

func getStdinData() string {
	stat, _ := os.Stdin.Stat()

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(data))
	}
	return ""
}

func runEditor(file string) error {
	// Retrieve the $EDITOR environment variable.
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return errors.New("$EDITOR is not set")
	}

	// Find the full path to the editor.
	path, err := exec.LookPath(editor)
	if err != nil {
		return fmt.Errorf("editor not found: %v", err)
	}

	// Build the argument slice; the first argument should be the command itself.
	args := []string{editor, file}

	// Load the current environment.
	env := os.Environ()

	// Replace current process with the editor using syscall.Exec.
	if err := syscall.Exec(path, args, env); err != nil {
		return fmt.Errorf("syscall.Exec failed: %v", err)
	}

	return nil
}
