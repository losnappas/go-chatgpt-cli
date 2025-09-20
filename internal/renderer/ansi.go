package renderer

import (
	"fmt"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

type Printer interface {
	Print(string)
	Close()
}

type Ansi struct {
	renderer *glamour.TermRenderer
	current  string
	started  bool
	output   *termenv.Output
	out      string
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

	output := termenv.DefaultOutput()
	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(getStyle(output.HasDarkBackground())),
		glamour.WithWordWrap(0),
	)
	if err != nil {
		return nil, err
	}

	return &Ansi{
		renderer: r,
		current:  "",
		output:   output,
	}, nil
}

func (r *Ansi) Close() {
	r.output.ExitAltScreen()
	fmt.Print(r.out)
}

func (r *Ansi) Print(text string) {
	// Append new text to the current buffer and re-render the whole thing
	r.current += text
	out, err := r.renderer.Render(r.current)
	if err != nil {
		panic(err)
	}
	if !r.started {
		r.output.AltScreen()
		r.output.SaveCursorPosition()
		r.started = true
	} else {
		r.output.ClearScreen()
		r.output.RestoreCursorPosition()
	}
	r.out = out
	fmt.Print(out)
}

func stringPtr(s string) *string {
	return &s
}

func getStyle(dark bool) ansi.StyleConfig {
	var s ansi.StyleConfig
	if dark {
		s = styles.DarkStyleConfig
		s.Document = ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("252"),
			},
		}
	} else {
		s = styles.LightStyleConfig
		s.Document = ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("234"),
			},
		}
	}
	return s
}
