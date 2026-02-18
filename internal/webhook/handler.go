package webhook

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sync/singleflight"
	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/store"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/verify"
)

// ociPullFunc is a package-level variable for test injection.
var ociPullFunc = store.PullOCI

const maxBodyBytes = 10 * 1024 * 1024 // 10 MB

// Handler returns an http.Handler that processes AdmissionReview requests.
func Handler(cfg Config) http.Handler {
	cache := newVerifierCache(time.Duration(cfg.CacheTTLSeconds) * time.Second)
	group := &singleflight.Group{}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleAdmission(w, r, cfg, cache, group)
	})
}

// HealthHandler returns an HTTP handler for liveness and readiness probes.
func HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func handleAdmission(w http.ResponseWriter, r *http.Request, cfg Config, cache *verifierCache, group *singleflight.Group) {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes))
	if err != nil {
		writeError(w, cfg, nil, fmt.Errorf("read body: %w", err))
		return
	}

	var review admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &review); err != nil {
		writeError(w, cfg, nil, fmt.Errorf("decode admission review: %w", err))
		return
	}
	if review.Request == nil {
		writeError(w, cfg, nil, fmt.Errorf("admission review has no request"))
		return
	}

	spec, err := podSpecFromResource(review.Request.Object.Raw)
	if err != nil {
		writeError(w, cfg, &review.Request.UID, fmt.Errorf("extract pod spec: %w", err))
		return
	}

	refs := ExtractImageRefs(*spec)
	var violations []string
	for _, ref := range refs {
		if err := verifyImage(ref, cfg, cache, group); err != nil {
			violations = append(violations, fmt.Sprintf("container %q (%s): %v", ref.Container, ref.Image, err))
		}
	}

	if len(violations) > 0 && !cfg.FailOpen {
		writeResponse(w, review.Request.UID, false, fmt.Sprintf("attestation verification failed: %v", violations))
		return
	}
	writeResponse(w, review.Request.UID, true, "all attestations verified")
}

func verifyImage(ref ImageRef, cfg Config, cache *verifierCache, group *singleflight.Group) error {
	ociRef, err := AttestationRef(cfg.RegistryPrefix, ref.Image)
	if err != nil {
		return fmt.Errorf("construct attestation ref: %w", err)
	}

	now := time.Now()
	if cache.hasFresh(ociRef, now) {
		return nil
	}
	run := func() error {
		// Re-check cache in case another in-flight request already populated it.
		if cache.hasFresh(ociRef, time.Now()) {
			return nil
		}
		if err := verifyImageNoCache(ociRef, cfg); err != nil {
			return err
		}
		cache.putSuccess(ociRef, time.Now())
		return nil
	}
	if group != nil {
		_, err, _ := group.Do(ociRef, func() (any, error) {
			return nil, run()
		})
		return err
	}
	return run()
}

func verifyImageNoCache(ociRef string, cfg Config) error {
	tmpDir, err := os.MkdirTemp("", "llmsa-webhook-")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	outPath := filepath.Join(tmpDir, "bundle.bundle.json")
	if err := ociPullFunc(ociRef, outPath); err != nil {
		return fmt.Errorf("pull attestation bundle: %w", err)
	}

	report := verify.Run(verify.Options{
		SourcePath: tmpDir,
		SchemaDir:  cfg.SchemaDir,
	})
	if !report.Passed {
		return fmt.Errorf("exit %d: %v", report.ExitCode, report.Violations)
	}
	return nil
}

// podSpecFromResource extracts the PodSpec from the raw object in the
// admission request. Supports Pod, Deployment, ReplicaSet, StatefulSet,
// DaemonSet, and Job resources.
func podSpecFromResource(raw []byte) (*corev1.PodSpec, error) {
	// Try Pod
	var pod corev1.Pod
	if err := json.Unmarshal(raw, &pod); err == nil && pod.Kind == "Pod" {
		return &pod.Spec, nil
	}
	// Try Deployment
	var deploy appsv1.Deployment
	if err := json.Unmarshal(raw, &deploy); err == nil && deploy.Kind == "Deployment" {
		return &deploy.Spec.Template.Spec, nil
	}
	// Try ReplicaSet
	var rs appsv1.ReplicaSet
	if err := json.Unmarshal(raw, &rs); err == nil && rs.Kind == "ReplicaSet" {
		return &rs.Spec.Template.Spec, nil
	}
	// Try StatefulSet
	var ss appsv1.StatefulSet
	if err := json.Unmarshal(raw, &ss); err == nil && ss.Kind == "StatefulSet" {
		return &ss.Spec.Template.Spec, nil
	}
	// Try DaemonSet
	var ds appsv1.DaemonSet
	if err := json.Unmarshal(raw, &ds); err == nil && ds.Kind == "DaemonSet" {
		return &ds.Spec.Template.Spec, nil
	}
	// Try Job
	var job batchv1.Job
	if err := json.Unmarshal(raw, &job); err == nil && job.Kind == "Job" {
		return &job.Spec.Template.Spec, nil
	}
	return nil, fmt.Errorf("unsupported resource kind")
}

func writeResponse(w http.ResponseWriter, uid k8stypes.UID, allowed bool, message string) {
	resp := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{APIVersion: "admission.k8s.io/v1", Kind: "AdmissionReview"},
		Response: &admissionv1.AdmissionResponse{
			UID:     uid,
			Allowed: allowed,
			Result:  &metav1.Status{Message: message},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, cfg Config, uid *k8stypes.UID, err error) {
	if cfg.FailOpen {
		respUID := k8stypes.UID("")
		if uid != nil {
			respUID = *uid
		}
		writeResponse(w, respUID, true, fmt.Sprintf("fail-open: %v", err))
		return
	}
	http.Error(w, err.Error(), http.StatusBadRequest)
}
