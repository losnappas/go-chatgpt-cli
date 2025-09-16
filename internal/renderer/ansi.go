package renderer

import (
	"bufio"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"golang.org/x/term"
)

type Printer interface {
	Print(string)
	Close()
}

type Ansi struct {
	renderer *glamour.TermRenderer
	current  string
	writer   *bufio.Writer
	viewport *tea.Program
}

type Plain struct {
}

func (r *Plain) Print(s string) {
	fmt.Print(s)
}

func (r *Plain) Close() {}

func NewPrinter() (Printer, error) {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return &Plain{}, nil
	}

	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil, err
	}

	p := tea.NewProgram(newViewport())
	go p.Run()
	return &Ansi{
		renderer: r,
		current:  "",
		viewport: p,
	}, nil
}

func (r *Ansi) Close() {
	r.viewport.Quit()
	r.viewport.Wait()
}

func (r *Ansi) Print(text string) {
	// Append new text to the current buffer and re-render the whole thing
	r.current += text
	out, err := r.renderer.Render(r.current)
	if err != nil {
		panic(err)
	}
	r.viewport.Send(newContent{out})
}
