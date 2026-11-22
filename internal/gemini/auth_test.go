package gemini

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestCookieCache_EmptyFile(t *testing.T) {
	cache := NewCookieCache("")
	info, err := cache.Load()
	if err != nil {
		t.Fatalf("expected no error for empty file path, got %v", err)
	}
	if info.Cookie != "" {
		t.Errorf("expected empty cookie, got %q", info.Cookie)
	}
}

func TestCookieCache_CachingAndTTL(t *testing.T) {
	tmpDir := t.TempDir()
	cookiePath := filepath.Join(tmpDir, "cookie.txt")

	// Write initial cookie
	if err := os.WriteFile(cookiePath, []byte("cookie_v1"), 0644); err != nil {
		t.Fatalf("failed to write temp cookie file: %v", err)
	}

	cache := NewCookieCache(cookiePath)

	// First load: hits disk
	info1, err := cache.Load()
	if err != nil {
		t.Fatalf("first Load failed: %v", err)
	}
	if info1.Cookie != "cookie_v1" {
		t.Fatalf("expected cookie_v1, got %q", info1.Cookie)
	}

	// Immediately overwrite file on disk
	if err := os.WriteFile(cookiePath, []byte("cookie_v2"), 0644); err != nil {
		t.Fatalf("failed to update temp cookie file: %v", err)
	}

	// Second load within 3 seconds: should hit fast-path cache and return cookie_v1
	info2, err := cache.Load()
	if err != nil {
		t.Fatalf("second Load failed: %v", err)
	}
	if info2.Cookie != "cookie_v1" {
		t.Errorf("expected cached cookie_v1, got %q", info2.Cookie)
	}

	// Manually age lastStat to simulate elapsed TTL
	cache.mu.Lock()
	cache.lastStat = time.Now().Add(-4 * time.Second)
	cache.mu.Unlock()

	// Third load after TTL: should reload from disk and return cookie_v2
	info3, err := cache.Load()
	if err != nil {
		t.Fatalf("third Load failed: %v", err)
	}
	if info3.Cookie != "cookie_v2" {
		t.Errorf("expected reloaded cookie_v2, got %q", info3.Cookie)
	}
}

func TestCookieCache_ConcurrentLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cookiePath := filepath.Join(tmpDir, "cookie.txt")

	if err := os.WriteFile(cookiePath, []byte("SAPISID=test_sapisid_value; SID=123"), 0644); err != nil {
		t.Fatalf("failed to write cookie file: %v", err)
	}

	cache := NewCookieCache(cookiePath)

	const concurrency = 50
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			info, err := cache.Load()
			if err != nil {
				t.Errorf("concurrent Load failed: %v", err)
				return
			}
			if info.SAPISID != "test_sapisid_value" {
				t.Errorf("expected SAPISID test_sapisid_value, got %q", info.SAPISID)
			}
		}()
	}

	wg.Wait()
}
