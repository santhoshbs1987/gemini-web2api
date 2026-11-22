package gemini

import (
	"strings"
	"testing"

	"github.com/ikhsan3adi/gemini-web2api/internal/config"
)

func TestApplyPersistenceFlagsTemporary(t *testing.T) {
	cfg := config.Config{TemporaryChats: true}
	inner := make([]any, GeminiPayloadSize)
	applyPersistenceFlags(inner, cfg)

	if inner[41] == nil {
		t.Fatal("inner[41] should not be nil")
	}
	flag, ok := inner[41].([]any)
	if !ok {
		t.Fatalf("inner[41] should be []any, got %T", inner[41])
	}
	if len(flag) != 1 || flag[0] != 1 {
		t.Errorf("inner[41] = %v, want [1]", flag)
	}
	if inner[45] != 1 {
		t.Errorf("inner[45] = %v, want 1", inner[45])
	}
}

func TestApplyPersistenceFlagsPersistent(t *testing.T) {
	cfg := config.Config{TemporaryChats: false}
	inner := make([]any, GeminiPayloadSize)
	applyPersistenceFlags(inner, cfg)

	flag, ok := inner[41].([]any)
	if !ok {
		t.Fatalf("inner[41] should be []any, got %T", inner[41])
	}
	if len(flag) != 1 || flag[0] != 2 {
		t.Errorf("inner[41] = %v, want [2]", flag)
	}
}

func TestBuildURLWithBLOverride(t *testing.T) {
	cfg := config.Config{
		AuthUser: "",
		GeminiBL: "old-bl",
	}
	url1 := BuildURL(cfg, "")
	if !strings.Contains(url1, "bl=old-bl") {
		t.Errorf("BuildURL with empty override should use cfg.GeminiBL, got %s", url1)
	}

	url2 := BuildURL(cfg, "new-bl")
	if !strings.Contains(url2, "bl=new-bl") {
		t.Errorf("BuildURL with override should use override, got %s", url2)
	}
}

func TestUUIDV4_FormatAndUniqueness(t *testing.T) {
	seen := make(map[string]bool, 1000)
	for i := 0; i < 1000; i++ {
		u := uuidV4()
		if len(u) != 36 {
			t.Fatalf("uuidV4() length = %d, want 36", len(u))
		}
		if u[8] != '-' || u[13] != '-' || u[18] != '-' || u[23] != '-' {
			t.Fatalf("uuidV4() malformed dashes: %s", u)
		}
		// Version 4 check
		if u[14] != '4' {
			t.Fatalf("uuidV4() version byte = %c, want 4", u[14])
		}
		// Variant 10xx check (8, 9, a, or b in hex)
		variant := u[19]
		if variant != '8' && variant != '9' && variant != 'a' && variant != 'b' {
			t.Fatalf("uuidV4() variant byte = %c, want 8, 9, a, or b", variant)
		}
		if seen[u] {
			t.Fatalf("uuidV4() generated collision: %s", u)
		}
		seen[u] = true
	}
}

func TestUpperUUIDV4_FormatAndUniqueness(t *testing.T) {
	seen := make(map[string]bool, 1000)
	for i := 0; i < 1000; i++ {
		u := upperUUIDV4()
		if len(u) != 36 {
			t.Fatalf("upperUUIDV4() length = %d, want 36", len(u))
		}
		if u[8] != '-' || u[13] != '-' || u[18] != '-' || u[23] != '-' {
			t.Fatalf("upperUUIDV4() malformed dashes: %s", u)
		}
		if u[14] != '4' {
			t.Fatalf("upperUUIDV4() version byte = %c, want 4", u[14])
		}
		variant := u[19]
		if variant != '8' && variant != '9' && variant != 'A' && variant != 'B' {
			t.Fatalf("upperUUIDV4() variant byte = %c, want 8, 9, A, or B", variant)
		}
		// Check that all letters are uppercase
		for _, ch := range u {
			if ch >= 'a' && ch <= 'z' {
				t.Fatalf("upperUUIDV4() contains lowercase char: %s", u)
			}
		}
		if seen[u] {
			t.Fatalf("upperUUIDV4() generated collision: %s", u)
		}
		seen[u] = true
	}
}
