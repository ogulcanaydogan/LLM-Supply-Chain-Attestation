package store

import (
	"fmt"
	"io"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

const bundleMediaType = types.MediaType("application/vnd.llmsa.bundle.v1+json")

func PublishOCI(inPath string, ociRef string) error {
	raw, err := os.ReadFile(inPath)
	if err != nil {
		return fmt.Errorf("read bundle: %w", err)
	}
	ref, err := name.ParseReference(ociRef, name.WithDefaultRegistry("ghcr.io"))
	if err != nil {
		return fmt.Errorf("parse oci ref: %w", err)
	}

	layer := static.NewLayer(raw, bundleMediaType)
	img, err := mutate.AppendLayers(empty.Image, layer)
	if err != nil {
		return fmt.Errorf("append layer: %w", err)
	}
	img = mutate.MediaType(img, types.OCIManifestSchema1)

	if err := remote.Write(ref, img, remote.WithAuthFromKeychain(authn.DefaultKeychain)); err != nil {
		return fmt.Errorf("push oci artifact: %w", err)
	}
	return nil
}

func PullOCI(ociRef string, outPath string) error {
	ref, err := name.ParseReference(ociRef, name.WithDefaultRegistry("ghcr.io"))
	if err != nil {
		return fmt.Errorf("parse oci ref: %w", err)
	}
	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return fmt.Errorf("pull oci artifact: %w", err)
	}
	layers, err := img.Layers()
	if err != nil {
		return fmt.Errorf("read layers: %w", err)
	}
	if len(layers) == 0 {
		return fmt.Errorf("oci artifact has no layers")
	}

	rc, err := layers[0].Uncompressed()
	if err != nil {
		return fmt.Errorf("read layer payload: %w", err)
	}
	defer rc.Close()
	raw, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("read layer bytes: %w", err)
	}
	if err := os.WriteFile(outPath, raw, 0o644); err != nil {
		return fmt.Errorf("write pulled bundle: %w", err)
	}
	return nil
}
