package main

import (
	"strings"
	"testing"
	"time"
)

func TestParseGeminiSKU(t *testing.T) {
	tests := []struct {
		description, model, direction string
	}{
		{"Generate content input token count gemini 2.5 flash lite short input text", "gemini-2.5-flash-lite", "input"},
		{"Generate content output token count gemini 3.5 flash lite text", "gemini-3.5-flash-lite", "output"},
		{"Generate content input token count gemini 3.1 flash lite preview text", "gemini-3.1-flash-lite-preview", "input"},
		{"Generate content output token count gemini 2.5 flash lite short output text non-thinking", "gemini-2.5-flash-lite", "output"},
		{"Generate content output token count gemini 3.8 flash text", "gemini-3.8-flash", "output"},
		{"Search grounding requests", "Search grounding requests", ""},
	}
	for _, tt := range tests {
		model, direction := parseGeminiSKU(tt.description)
		if model != tt.model || direction != tt.direction {
			t.Errorf("parseGeminiSKU(%q) = %q, %q, want %q, %q", tt.description, model, direction, tt.model, tt.direction)
		}
	}
}

func TestLineItemModel(t *testing.T) {
	tests := map[string]string{
		"gpt-4o-2024-08-06, input":             "gpt-4o-2024-08-06",
		"gpt-4o-mini-2024-07-18, cached input": "gpt-4o-mini-2024-07-18",
		"web search tool calls":                "web search tool calls",
		"":                                     "unknown",
	}
	for lineItem, want := range tests {
		if got := lineItemModel(lineItem); got != want {
			t.Errorf("lineItemModel(%q) = %q, want %q", lineItem, got, want)
		}
	}
}

func testReport() *Report {
	r := newReport("test", time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC), 2)
	r.HasRequests, r.HasTokens = true, true
	r.add("2026-11-20", "model-a", Stats{Cost: 1.5, Requests: 10, InputTokens: 1000, OutputTokens: 100})
	r.add("2026-11-20", "model-b", Stats{Cost: 0.25, Requests: 2, InputTokens: 50, OutputTokens: 5})
	r.add("2026-11-21", "model-b", Stats{Cost: 2, Requests: 3, InputTokens: 70, OutputTokens: 7})
	return r
}

func TestPrintColumns(t *testing.T) {
	other := newReport("other", time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC), 2)
	other.HasTokens = true // like gemini: tokens without requests
	other.add("2026-11-21", "", Stats{Cost: 12.5, InputTokens: 500, OutputTokens: 50})

	tests := []struct {
		name                      string
		verbose, showTotal, chart bool
		want                      string
	}{
		{"total", false, true, false, `
date       ┊ test   ┊ other
┈┈┈┈┈┈┈┈┈┈┈┼┈┈┈┈┈┈┈┈┼┈┈┈┈┈┈┈┈
2026-11-20 ┊ 1.7500 ┊  0.0000
2026-11-21 ┊ 2.0000 ┊ 12.5000
Total:     ┊ 3.7500 ┊ 12.5000
`},
		{"verbose", true, true, false, `
date       ┊ test    req    in  out ┊ other     in  out
` + strings.Repeat("┈", 11) + "┼" + strings.Repeat("┈", 24) + "┼" + strings.Repeat("┈", 18) + `
2026-11-20 ┊ 1.7500   12  1050  105 ┊  0.0000    0    0
2026-11-21 ┊ 2.0000    3    70    7 ┊ 12.5000  500   50
Total:     ┊ 3.7500   15  1120  112 ┊ 12.5000  500   50
`},
		{"chart", false, false, true, `
date       ┊ test      ┊ other
┈┈┈┈┈┈┈┈┈┈┈┼┈┈┈┈┈┈┈┈┈┈┈┼┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈
2026-11-20 ┊ 1.7500 🬋🬃 ┊  0.0000
2026-11-21 ┊ 2.0000 🬋🬃 ┊ 12.5000 🬋🬋🬋🬋🬋🬋🬋🬋🬋🬋
`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out strings.Builder
			printColumns(&out, []*Report{testReport(), other}, tt.verbose, tt.showTotal, tt.chart)
			if want := strings.TrimPrefix(tt.want, "\n"); out.String() != want {
				t.Errorf("got:\n%s\nwant:\n%s", out.String(), want)
			}
		})
	}
}

func TestPrintReport(t *testing.T) {
	tests := []struct {
		name                    string
		byModel, verbose, chart bool
		want                    string
	}{
		{"default", false, false, false, `
2026-11-20 1.7500
2026-11-21 2.0000
Total: 3.7500
`},
		{"default-chart", false, false, true, `
2026-11-20 1.7500 🬋🬋🬋🬋🬋🬋🬋🬋🬋
2026-11-21 2.0000 🬋🬋🬋🬋🬋🬋🬋🬋🬋🬋
Total: 3.7500
`},
		{"by-model-chart", true, false, true, `
2026-11-20  1.7500  🬋🬋🬋🬋🬋🬋🬋🬋🬋
  model-a   1.5000  🬋🬋🬋🬋🬋🬋🬋🬃
  model-b   0.2500  🬋🬃
2026-11-21  2.0000  🬋🬋🬋🬋🬋🬋🬋🬋🬋🬋
  model-b   2.0000  🬋🬋🬋🬋🬋🬋🬋🬋🬋🬋
Total:      3.7500
  model-b   2.2500  🬋🬋🬋🬋🬋🬋
  model-a   1.5000  🬋🬋🬋🬋
`},
		{"by-model", true, false, false, `
2026-11-20  1.7500
  model-a   1.5000
  model-b   0.2500
2026-11-21  2.0000
  model-b   2.0000
Total:      3.7500
  model-b   2.2500
  model-a   1.5000
`},
		{"verbose", false, true, false, `
date          cost  requests  input_tokens  output_tokens
2026-11-20  1.7500        12          1050            105
2026-11-21  2.0000         3            70              7
Total:      3.7500        15          1120            112
`},
		{"by-model-verbose", true, true, false, `
date          cost  requests  input_tokens  output_tokens
2026-11-20  1.7500        12          1050            105
  model-a   1.5000        10          1000            100
  model-b   0.2500         2            50              5
2026-11-21  2.0000         3            70              7
  model-b   2.0000         3            70              7
Total:      3.7500        15          1120            112
  model-b   2.2500         5           120             12
  model-a   1.5000        10          1000            100
`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out strings.Builder
			printReport(&out, testReport(), tt.byModel, tt.verbose, true, tt.chart)
			if want := strings.TrimPrefix(tt.want, "\n"); out.String() != want {
				t.Errorf("got:\n%s\nwant:\n%s", out.String(), want)
			}
		})
	}
}
