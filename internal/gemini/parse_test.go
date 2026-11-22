package gemini

import (
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
