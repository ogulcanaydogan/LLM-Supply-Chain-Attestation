package webhook

import (
	"testing"
	"time"
)

func TestDefaultConfigIncludesCacheTTL(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.CacheTTLSeconds <= 0 {
		t.Fatalf("expected positive cache ttl, got %d", cfg.CacheTTLSeconds)
	}
}

func TestVerifierCacheFreshAndExpiry(t *testing.T) {
	cache := newVerifierCache(100 * time.Millisecond)
	if cache == nil {
		t.Fatal("expected cache instance")
	}
	now := time.Now()
	key := "ghcr.io/acme/attestations:sha256-deadbeef"
	if cache.hasFresh(key, now) {
		t.Fatalf("unexpected cache hit before put")
	}
	cache.putSuccess(key, now)
	if !cache.hasFresh(key, now.Add(50*time.Millisecond)) {
		t.Fatalf("expected cache hit before expiry")
	}
	if cache.hasFresh(key, now.Add(200*time.Millisecond)) {
		t.Fatalf("expected cache miss after expiry")
	}
}

func TestVerifierCacheDisabled(t *testing.T) {
	if cache := newVerifierCache(0); cache != nil {
		t.Fatalf("expected nil cache when ttl disabled")
	}
}
