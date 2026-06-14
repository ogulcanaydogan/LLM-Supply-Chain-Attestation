package webhook

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sync/singleflight"
)

// pullCounter returns an ociPullFunc replacement that copies a valid bundle and
// records how many times it was invoked, optionally sleeping to widen the window
// in which concurrent callers overlap.
func pullCounter(t *testing.T, bundleDir string, counter *int32, delay time.Duration) func(string, string) error {
	t.Helper()
	return func(_ string, outPath string) error {
		atomic.AddInt32(counter, 1)
		if delay > 0 {
			time.Sleep(delay)
		}
		data, err := os.ReadFile(filepath.Join(bundleDir, "bundle.bundle.json"))
		if err != nil {
			return err
		}
		return os.WriteFile(outPath, data, 0o644)
	}
}

func cacheTestConfig() Config {
	return Config{
		RegistryPrefix:  "ghcr.io/test/attestations",
		SchemaDir:       "../../schemas/v1",
		CacheTTLSeconds: 60,
	}
}

// TestVerifyImageConcurrentDedup asserts that singleflight collapses many
// concurrent identical verifications into a single OCI pull.
func TestVerifyImageConcurrentDedup(t *testing.T) {
	bundleDir := t.TempDir()
	writeValidBundle(t, bundleDir)

	var pulls int32
	original := ociPullFunc
	ociPullFunc = pullCounter(t, bundleDir, &pulls, 25*time.Millisecond)
	t.Cleanup(func() { ociPullFunc = original })

	cfg := cacheTestConfig()
	cache := newVerifierCache(time.Duration(cfg.CacheTTLSeconds) * time.Second)
	group := &singleflight.Group{}
	ref := ImageRef{Container: "app", Image: "myapp@sha256:abc123"}

	const n = 24
	var wg sync.WaitGroup
	errs := make([]error, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			errs[idx] = verifyImage(ref, cfg, cache, group)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d: unexpected error: %v", i, err)
		}
	}
	if got := atomic.LoadInt32(&pulls); got != 1 {
		t.Fatalf("expected exactly 1 OCI pull for %d concurrent identical requests, got %d", n, got)
	}
}

// TestVerifyImageWarmCacheSkipsPull asserts that a warm cache short-circuits the
// pull/verify path entirely on subsequent calls.
func TestVerifyImageWarmCacheSkipsPull(t *testing.T) {
	bundleDir := t.TempDir()
	writeValidBundle(t, bundleDir)

	var pulls int32
	original := ociPullFunc
	ociPullFunc = pullCounter(t, bundleDir, &pulls, 0)
	t.Cleanup(func() { ociPullFunc = original })

	cfg := cacheTestConfig()
	cache := newVerifierCache(time.Duration(cfg.CacheTTLSeconds) * time.Second)
	group := &singleflight.Group{}
	ref := ImageRef{Container: "app", Image: "myapp@sha256:abc123"}

	// Cold call populates the cache.
	if err := verifyImage(ref, cfg, cache, group); err != nil {
		t.Fatalf("cold verify: %v", err)
	}
	if got := atomic.LoadInt32(&pulls); got != 1 {
		t.Fatalf("expected 1 pull after cold call, got %d", got)
	}

	// Subsequent warm calls must not pull again.
	for i := 0; i < 5; i++ {
		if err := verifyImage(ref, cfg, cache, group); err != nil {
			t.Fatalf("warm verify %d: %v", i, err)
		}
	}
	if got := atomic.LoadInt32(&pulls); got != 1 {
		t.Fatalf("warm cache should not trigger more pulls, got %d total", got)
	}
}

// TestVerifyImageWarmCacheLatency asserts the warm-cache path is materially
// faster than a cold verification that performs a (simulated) slow pull.
func TestVerifyImageWarmCacheLatency(t *testing.T) {
	bundleDir := t.TempDir()
	writeValidBundle(t, bundleDir)

	var pulls int32
	original := ociPullFunc
	ociPullFunc = pullCounter(t, bundleDir, &pulls, 50*time.Millisecond)
	t.Cleanup(func() { ociPullFunc = original })

	cfg := cacheTestConfig()
	cache := newVerifierCache(time.Duration(cfg.CacheTTLSeconds) * time.Second)
	group := &singleflight.Group{}
	ref := ImageRef{Container: "app", Image: "myapp@sha256:abc123"}

	coldStart := time.Now()
	if err := verifyImage(ref, cfg, cache, group); err != nil {
		t.Fatalf("cold verify: %v", err)
	}
	coldDuration := time.Since(coldStart)

	warmStart := time.Now()
	if err := verifyImage(ref, cfg, cache, group); err != nil {
		t.Fatalf("warm verify: %v", err)
	}
	warmDuration := time.Since(warmStart)

	if warmDuration >= coldDuration {
		t.Fatalf("expected warm cache faster than cold (cold=%s warm=%s)", coldDuration, warmDuration)
	}
}
