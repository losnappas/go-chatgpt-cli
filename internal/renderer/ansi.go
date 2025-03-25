package renderer

import (
	"fmt"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	styles "github.com/charmbracelet/glamour/styles"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

type Printer interface {
	Print(string)
}

type Ansi struct {
	renderer *glamour.TermRenderer
	current  string
}

type Plain struct {
}

func NewPrinter() (Printer, error) {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return &Plain{}, nil
	}

	style, err := getDefaultStyle("auto")
	if err != nil {
		return nil, err
	}
	// Removes margin and empty lines before/after.
	style.Document = ansi.StyleBlock{}
	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(*style),
		// Makes copy pasting the response more pleasant.
		glamour.WithWordWrap(0),
	)
	if err != nil {
		return nil, err
	}
	// Save cursor position with both legacy and CSI variants
	fmt.Print("\x1b7\x1b[s")
	return &Ansi{
		renderer: r,
		current:  "",
	}, nil
}

func (r *Plain) Print(s string) {
	fmt.Print(s)
}

// clearPrevious erases the previously rendered block from the terminal.
func (r *Ansi) clearPrevious() {
	// Restore saved cursor position (use both legacy ESC7/ESC8 and CSI s/u for broader support)
	// Then clear to end of screen so all previously printed content is removed.
	fmt.Print("\x1b8\x1b[u\x1b[J")
}

func (r *Ansi) Print(text string) {
	// Append new text to the current buffer and re-render the whole thing
	r.current += text
	out, err := r.renderer.Render(r.current)
	if err != nil {
		return
	}

	r.clearPrevious()

	fmt.Print(out)
}

// getDefaultStyle is copy paste from glamour package.
func getDefaultStyle(style string) (*ansi.StyleConfig, error) {
	if style == styles.AutoStyle {
		if !term.IsTerminal(int(os.Stdout.Fd())) {
			return &styles.NoTTYStyleConfig, nil
		}
		if termenv.HasDarkBackground() {
			return &styles.DarkStyleConfig, nil
		}
		return &styles.LightStyleConfig, nil
	}

	styles, ok := styles.DefaultStyles[style]
	if !ok {
		return nil, fmt.Errorf("%s: style not found", style)
	}
	return styles, nil
}
