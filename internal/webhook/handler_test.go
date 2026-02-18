package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/sign"
)

// --- Extractor Tests ---

func TestExtractImageRefs(t *testing.T) {
	spec := corev1.PodSpec{
		InitContainers: []corev1.Container{
			{Name: "init", Image: "busybox:latest"},
		},
		Containers: []corev1.Container{
			{Name: "app", Image: "myapp:v1"},
			{Name: "sidecar", Image: "envoy:1.30"},
		},
		EphemeralContainers: []corev1.EphemeralContainer{
			{EphemeralContainerCommon: corev1.EphemeralContainerCommon{Name: "debug", Image: "busybox:debug"}},
		},
	}

	refs := ExtractImageRefs(spec)
	if len(refs) != 4 {
		t.Fatalf("expected 4 refs, got %d", len(refs))
	}
	if refs[0].Container != "init" || refs[0].Image != "busybox:latest" {
		t.Errorf("init container: got %+v", refs[0])
	}
	if refs[1].Container != "app" || refs[1].Image != "myapp:v1" {
		t.Errorf("app container: got %+v", refs[1])
	}
	if refs[3].Container != "debug" {
		t.Errorf("ephemeral container: got %+v", refs[3])
	}
}

func TestExtractImageRefsEmpty(t *testing.T) {
	refs := ExtractImageRefs(corev1.PodSpec{})
	if len(refs) != 0 {
		t.Fatalf("expected 0 refs, got %d", len(refs))
	}
}

func TestAttestationRef(t *testing.T) {
	ref, err := AttestationRef("ghcr.io/org/attestations", "nginx@sha256:abcdef1234567890")
	if err != nil {
		t.Fatal(err)
	}
	if ref != "ghcr.io/org/attestations:sha256-abcdef1234567890" {
		t.Errorf("ref = %q", ref)
	}
}

func TestAttestationRef_Tagged(t *testing.T) {
	ref, err := AttestationRef("ghcr.io/org/attestations", "nginx:1.25")
	if err != nil {
		t.Fatal(err)
	}
	if ref == "" {
		t.Error("expected non-empty ref for tagged image")
	}
}

func TestAttestationRef_EmptyPrefix(t *testing.T) {
	_, err := AttestationRef("", "nginx:1.25")
	if err == nil {
		t.Fatal("expected error for empty prefix")
	}
}

// --- Handler Tests ---

func writeValidBundle(t testing.TB, dir string) {
	t.Helper()

	// Generate a real PEM key and sign properly so verify.Run passes.
	keyPath := filepath.Join(dir, "key.pem")
	if err := sign.GeneratePEMPrivateKey(keyPath); err != nil {
		t.Fatal(err)
	}
	signer, err := sign.NewPEMSigner(keyPath)
	if err != nil {
		t.Fatal(err)
	}

	statement := map[string]any{
		"schema_version":   "1.0.0",
		"statement_id":     "stmt-test",
		"attestation_type": "prompt_attestation",
		"predicate_type":   "https://llmsa.dev/attestation/prompt/v1",
		"generated_at":     time.Now().UTC().Format(time.RFC3339),
		"generator": map[string]any{
			"name": "llmsa", "version": "0.1.0", "git_sha": "abc123",
		},
		"subject": []any{},
		"predicate": map[string]any{
			"prompt_bundle_digest": "sha256:abc",
			"system_prompt_digest": "sha256:def",
			"template_digests":     []any{"sha256:t1"},
			"tool_schema_digests":  []any{"sha256:s1"},
			"safety_policy_digest": "sha256:safety",
		},
		"privacy": map[string]any{"mode": "hash_only"},
	}

	// Use the same canonical JSON that CreateBundle uses internally.
	canonical, err := hash.CanonicalJSON(statement)
	if err != nil {
		t.Fatal(err)
	}
	material, err := signer.Sign(canonical)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := sign.CreateBundle(statement, material)
	if err != nil {
		t.Fatal(err)
	}
	if err := sign.WriteBundle(filepath.Join(dir, "bundle.bundle.json"), bundle); err != nil {
		t.Fatal(err)
	}
}

func buildAdmissionReview(t testing.TB, obj any) []byte {
	t.Helper()
	raw, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	review := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{APIVersion: "admission.k8s.io/v1", Kind: "AdmissionReview"},
		Request: &admissionv1.AdmissionRequest{
			UID: types.UID("test-uid"),
			Object: runtime.RawExtension{
				Raw: raw,
			},
		},
	}
	data, err := json.Marshal(review)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestHandlerAllowValidAttestation(t *testing.T) {
	bundleDir := t.TempDir()
	writeValidBundle(t, bundleDir)

	original := ociPullFunc
	ociPullFunc = func(ociRef, outPath string) error {
		data, err := os.ReadFile(filepath.Join(bundleDir, "bundle.bundle.json"))
		if err != nil {
			return err
		}
		return os.WriteFile(outPath, data, 0o644)
	}
	t.Cleanup(func() { ociPullFunc = original })

	cfg := Config{
		RegistryPrefix: "ghcr.io/test/attestations",
		SchemaDir:      "../../schemas/v1",
	}

	pod := corev1.Pod{
		TypeMeta: metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		Spec:     corev1.PodSpec{Containers: []corev1.Container{{Name: "app", Image: "myapp@sha256:abc123"}}},
	}
	body := buildAdmissionReview(t, pod)

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	Handler(cfg).ServeHTTP(rec, req)

	var resp admissionv1.AdmissionReview
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Response == nil {
		t.Fatal("response is nil")
	}
	if !resp.Response.Allowed {
		t.Errorf("expected allowed, got denied: %s", resp.Response.Result.Message)
	}
}

func TestHandlerCachesSuccessfulVerification(t *testing.T) {
	bundleDir := t.TempDir()
	writeValidBundle(t, bundleDir)

	original := ociPullFunc
	pullCount := 0
	ociPullFunc = func(_ string, outPath string) error {
		pullCount++
		data, err := os.ReadFile(filepath.Join(bundleDir, "bundle.bundle.json"))
		if err != nil {
			return err
		}
		return os.WriteFile(outPath, data, 0o644)
	}
	t.Cleanup(func() { ociPullFunc = original })

	cfg := Config{
		RegistryPrefix:  "ghcr.io/test/attestations",
		SchemaDir:       "../../schemas/v1",
		CacheTTLSeconds: 60,
	}
	handler := Handler(cfg)

	pod := corev1.Pod{
		TypeMeta: metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		Spec:     corev1.PodSpec{Containers: []corev1.Container{{Name: "app", Image: "myapp@sha256:abc123"}}},
	}
	body := buildAdmissionReview(t, pod)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		var resp admissionv1.AdmissionReview
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if resp.Response == nil || !resp.Response.Allowed {
			t.Fatalf("expected allowed response on iteration %d, got %#v", i, resp.Response)
		}
	}
	if pullCount != 1 {
		t.Fatalf("expected exactly one OCI pull with warm cache, got %d", pullCount)
	}
}

func TestHandlerDenyMissingAttestation(t *testing.T) {
	original := ociPullFunc
	ociPullFunc = func(_, _ string) error {
		return fmt.Errorf("not found")
	}
	t.Cleanup(func() { ociPullFunc = original })

	cfg := Config{
		RegistryPrefix: "ghcr.io/test/attestations",
		SchemaDir:      "../../schemas/v1",
	}

	pod := corev1.Pod{
		TypeMeta: metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		Spec:     corev1.PodSpec{Containers: []corev1.Container{{Name: "app", Image: "unattested:latest"}}},
	}
	body := buildAdmissionReview(t, pod)

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	Handler(cfg).ServeHTTP(rec, req)

	var resp admissionv1.AdmissionReview
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Response == nil {
		t.Fatal("response is nil")
	}
	if resp.Response.Allowed {
		t.Error("expected denied for missing attestation")
	}
}

func TestHandlerFailOpenOnError(t *testing.T) {
	original := ociPullFunc
	ociPullFunc = func(_, _ string) error {
		return fmt.Errorf("registry unavailable")
	}
	t.Cleanup(func() { ociPullFunc = original })

	cfg := Config{
		RegistryPrefix: "ghcr.io/test/attestations",
		SchemaDir:      "../../schemas/v1",
		FailOpen:       true,
	}

	pod := corev1.Pod{
		TypeMeta: metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		Spec:     corev1.PodSpec{Containers: []corev1.Container{{Name: "app", Image: "myapp:v1"}}},
	}
	body := buildAdmissionReview(t, pod)

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	Handler(cfg).ServeHTTP(rec, req)

	var resp admissionv1.AdmissionReview
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Response == nil {
		t.Fatal("response is nil")
	}
	if !resp.Response.Allowed {
		t.Error("expected allowed in fail-open mode")
	}
}

func TestHandlerDeploymentExtraction(t *testing.T) {
	original := ociPullFunc
	ociPullFunc = func(_, _ string) error {
		return fmt.Errorf("not found")
	}
	t.Cleanup(func() { ociPullFunc = original })

	cfg := Config{
		RegistryPrefix: "ghcr.io/test/attestations",
		SchemaDir:      "../../schemas/v1",
	}

	deploy := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "web", Image: "nginx:1.25"},
					},
				},
			},
		},
	}
	body := buildAdmissionReview(t, deploy)

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	Handler(cfg).ServeHTTP(rec, req)

	var resp admissionv1.AdmissionReview
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Response == nil {
		t.Fatal("response is nil")
	}
	// Should be denied because ociPullFunc returns error
	if resp.Response.Allowed {
		t.Error("expected denied for unattested deployment")
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	HealthHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("body = %q, want ok", rec.Body.String())
	}
}

func TestPodSpecFromResource_UnsupportedKind(t *testing.T) {
	raw, _ := json.Marshal(map[string]any{"kind": "ConfigMap"})
	_, err := podSpecFromResource(raw)
	if err == nil {
		t.Fatal("expected error for unsupported kind")
	}
}

func BenchmarkHandlerWarmCache(b *testing.B) {
	bundleDir := b.TempDir()
	writeValidBundle(b, bundleDir)

	original := ociPullFunc
	defer func() { ociPullFunc = original }()
	pullCount := 0
	ociPullFunc = func(_ string, outPath string) error {
		pullCount++
		data, err := os.ReadFile(filepath.Join(bundleDir, "bundle.bundle.json"))
		if err != nil {
			return err
		}
		return os.WriteFile(outPath, data, 0o644)
	}

	cfg := Config{
		RegistryPrefix:  "ghcr.io/test/attestations",
		SchemaDir:       "../../schemas/v1",
		CacheTTLSeconds: 120,
	}
	handler := Handler(cfg)
	pod := corev1.Pod{
		TypeMeta: metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		Spec:     corev1.PodSpec{Containers: []corev1.Container{{Name: "app", Image: "myapp@sha256:bench"}}},
	}
	body := buildAdmissionReview(b, pod)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status code %d", rec.Code)
		}
	}
	b.StopTimer()
	if pullCount > 1 {
		b.Fatalf("expected warm cache to avoid repeated pulls, pull count=%d", pullCount)
	}
}
