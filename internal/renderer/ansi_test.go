package renderer

import (
	"regexp"
	"testing"
)

func runCase(str string) {
	rend, _ := NewPrinter()
	re := regexp.MustCompile(`\s`)

	parts := re.Split(str, -1)
	separators := re.FindAllString(str, -1)

	var result []string
	for i, p := range parts {
		if p != "" {
			result = append(result, p)
		}
		if i < len(separators) {
			result = append(result, separators[i])
		}
	}
	for _, s := range result {
		rend.Print(s)
	}
}

func TestPrint(t *testing.T) {
	str := `Truly arbitrary-length video generation remains an open challenge. Current video models (e.g., OpenAI Sora, Pika, Runway Gen-2, Stability AI’s Stable Video Diffusion) typically produce short clips (a few seconds to under a minute).

For longer outputs, approaches include:
- **Looping or chaining clips** (sequential generation with temporal alignment).
- **Training recurrent/streaming transformer-based models** that can extend outputs, though stability degrades over time.
- **Research directions**: latent diffusion with temporal consistency modules and autoregressive frame prediction to enable longer continuities.

No widely available model today can *natively* generate arbitrarily long, coherent video without degradation or manual stitching.

Would you like me to list some of the most promising open-source projects you could experiment with for extended-length video?
`
	runCase(str)
}

func TestPrint2(t *testing.T) {
	str := `Received your message — I’m here and ready. What would you like to test?`
	runCase(str)
}

func TestPrint3(t *testing.T) {
	str := `**Received your message** — I’m here and ready. What would you like to test?

- perhaps
- me?`
	runCase(str)
}
