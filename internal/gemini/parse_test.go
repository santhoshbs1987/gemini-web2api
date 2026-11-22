package gemini

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCleanText(t *testing.T) {
	input := "Hello world\n```python?code_reference&code_event_index=1\nprint('hi')```\nhttp://googleusercontent.com/card_content/123\nExtra text"
	cleaned := CleanText(input, true)
	expected := "Hello world\nExtra text"
	if cleaned != expected {
		t.Errorf("CleanText got %q, want %q", cleaned, expected)
	}

	// Edge case 1: plain text without any backticks or card content (fast path)
	plain := "Just normal plain text without formatting."
	if CleanText(plain, false) != plain {
		t.Errorf("CleanText mutated plain text: got %q, want %q", CleanText(plain, false), plain)
	}

	// Edge case 2: normal markdown code block should NOT be stripped
	codeBlock := "Here is code:\n```go\nfmt.Println(\"hello\")\n```\nEnjoy!"
	if CleanText(codeBlock, false) != codeBlock {
		t.Errorf("CleanText stripped valid markdown code fence: got %q, want %q", CleanText(codeBlock, false), codeBlock)
	}

	// Edge case 3: empty string
	if CleanText("", false) != "" {
		t.Errorf("CleanText failed on empty string")
	}
}

func TestIsBardError(t *testing.T) {
	raw := "something BardErrorInfo [1024] happened"
	code, ok := IsBardError(raw)
	if !ok || code != "1024" {
		t.Errorf("IsBardError got (%q, %t), want (\"1024\", true)", code, ok)
	}

	// Fast-path: string without BardErrorInfo
	noErr := "200 OK standard response chunk"
	if _, ok := IsBardError(noErr); ok {
		t.Errorf("IsBardError should be false for normal string")
	}
}

func TestExtractTextsFromLine(t *testing.T) {
	inner := make([]any, 5)
	inner[0] = strings.Repeat("metadata_filler_string_", 10)
	inner[4] = []any{[]any{"c_1", []any{"Hello from Gemini!"}}}
	innerBytes, _ := json.Marshal(inner)
	outer := []any{[]any{"wrb.fr", nil, string(innerBytes)}}
	lineBytes, _ := json.Marshal(outer)

	texts := ExtractTextsFromLine(string(lineBytes))
	if len(texts) != 1 || texts[0] != "Hello from Gemini!" {
		t.Fatalf("ExtractTextsFromLine got %v, want ['Hello from Gemini!']", texts)
	}
}

func TestStreamParserFeed(t *testing.T) {
	parser := NewStreamParser()

	// Chunk 1
	inner1 := make([]any, 5)
	inner1[0] = strings.Repeat("filler_", 20)
	inner1[4] = []any{[]any{"c_1", []any{"Hello"}}}
	b1, _ := json.Marshal(inner1)
	o1, _ := json.Marshal([]any{[]any{"wrb.fr", nil, string(b1)}})

	deltas1, err := parser.Feed(string(o1) + "\n")
	if err != nil {
		t.Fatalf("Feed chunk 1 error: %v", err)
	}
	if len(deltas1) != 1 || deltas1[0] != "Hello" {
		t.Fatalf("Feed chunk 1 deltas = %v, want ['Hello']", deltas1)
	}

	// Chunk 2 (cumulative: "Hello world")
	inner2 := make([]any, 5)
	inner2[0] = strings.Repeat("filler_", 20)
	inner2[4] = []any{[]any{"c_1", []any{"Hello world"}}}
	b2, _ := json.Marshal(inner2)
	o2, _ := json.Marshal([]any{[]any{"wrb.fr", nil, string(b2)}})

	deltas2, err := parser.Feed(string(o2) + "\n")
	if err != nil {
		t.Fatalf("Feed chunk 2 error: %v", err)
	}
	if len(deltas2) != 1 || deltas2[0] != " world" {
		t.Fatalf("Feed chunk 2 deltas = %v, want [' world']", deltas2)
	}
}
