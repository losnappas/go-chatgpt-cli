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
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/losnappas/go-chatgpt-cli/internal/conversation"
	"github.com/losnappas/go-chatgpt-cli/internal/history"
	"github.com/losnappas/go-chatgpt-cli/internal/llm"
	"github.com/losnappas/go-chatgpt-cli/internal/renderer"
)

var (
	model       = flag.String("model", "openai/o3-mini", "The provider/model to use")
	apiKey      = flag.String("api-key", "", "The API key to use, as provider=api_key")
	historyFile = flag.String("history-file", "", "Read chat history from markdown file")
	clear       = flag.Bool("c", false, "Clear history")
	editor      = flag.Bool("editor", false, "Open history with $EDITOR")
	system      = flag.String("system-prompt", "", "The LLM system prompt. Overridden by history file")
)

func runCase(str string) {
	rend, err := renderer.NewPrinter()
	if err != nil {
		panic(err)
	}
	defer rend.Close()
	re := regexp.MustCompile(`\s+`)

	parts := re.Split(str, -1)
	separators := re.FindAllString(str, -1)

	var result []string
	for i, p := range parts {
		var sep string
		if i < len(separators) {
			sep = separators[i]
		}
		if p != "" {
			result = append(result, p+sep)
		}
	}
	for _, s := range result {
		rend.Print(s)
		time.Sleep(time.Millisecond * 100)
	}
}

func main() {
	str := `Truly arbitrary-length video generation remains an open challenge. Current video models (e.g., OpenAI Sora, Pika, Runway Gen-2, Stability AI’s Stable Video Diffusion) typically produce short clips (a few seconds to under a minute).

For longer outputs, approaches include:
- **Looping or chaining clips** (sequential generation with temporal alignment).
- **Training recurrent/streaming transformer-based models** that can extend outputs, though stability degrades over time.
- **Research directions**: latent diffusion with temporal consistency modules and autoregressive frame prediction to enable longer continuities.

No widely available model today can *natively* generate arbitrarily long, coherent video without degradation or manual stitching.

Would you like me to list some of the most promising open-source projects you could experiment with for extended-length video?`
	runCase(str)
	// main2()
}

func main2() {
	ctx := context.Background()
	flag.Parse()
	err := run(ctx)
	if err != nil {
		log.Fatalln(err)
	}
}

func run(ctx context.Context) error {
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
