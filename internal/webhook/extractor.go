package webhook

import (
	"crypto/sha256"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// ImageRef represents a container image reference extracted from a Pod spec.
type ImageRef struct {
	Container string
	Image     string
}

// ExtractImageRefs returns all container image references from a Pod spec.
// It scans initContainers, containers, and ephemeralContainers.
func ExtractImageRefs(spec corev1.PodSpec) []ImageRef {
	refs := make([]ImageRef, 0, len(spec.InitContainers)+len(spec.Containers))
	for _, c := range spec.InitContainers {
		refs = append(refs, ImageRef{Container: c.Name, Image: c.Image})
	}
	for _, c := range spec.Containers {
		refs = append(refs, ImageRef{Container: c.Name, Image: c.Image})
	}
	for _, c := range spec.EphemeralContainers {
		refs = append(refs, ImageRef{Container: c.Name, Image: c.Image})
	}
	return refs
}

// AttestationRef constructs the OCI reference where the attestation bundle
// for a given image is expected to reside within the registry prefix.
// For image "nginx@sha256:abc123", it produces "{prefix}:sha256-abc123".
// For image "nginx:1.25", it produces "{prefix}:nginx-1.25".
func AttestationRef(registryPrefix, imageRef string) (string, error) {
	if registryPrefix == "" {
		return "", fmt.Errorf("registry prefix is required")
	}
	tag := sanitiseImageTag(imageRef)
	if tag == "" {
		return "", fmt.Errorf("cannot derive attestation tag from image ref %q", imageRef)
	}
	return fmt.Sprintf("%s:%s", strings.TrimSuffix(registryPrefix, "/"), tag), nil
}

// sanitiseImageTag converts an image reference into a safe OCI tag.
func sanitiseImageTag(imageRef string) string {
	// If the image contains a digest, use it as the tag.
	if idx := strings.LastIndex(imageRef, "@sha256:"); idx >= 0 {
		digest := imageRef[idx+1:] // "sha256:abc..."
		return strings.ReplaceAll(digest, ":", "-")
	}
	// Otherwise use a hash of the full reference for uniqueness.
	h := sha256.Sum256([]byte(imageRef))
	return fmt.Sprintf("img-%x", h[:8])
}
