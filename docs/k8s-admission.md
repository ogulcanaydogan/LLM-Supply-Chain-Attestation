# Kubernetes Validating Admission Webhook

The `llmsa webhook serve` command runs a Kubernetes validating admission webhook that enforces attestation verification at pod admission time. Every container image in a Pod, Deployment, ReplicaSet, StatefulSet, DaemonSet, or Job must have a valid attestation bundle in an OCI registry before the resource is admitted.

## Architecture

```
kubectl apply → K8s API Server → AdmissionReview → llmsa webhook
                                                       ↓
                                              Extract image refs
                                                       ↓
                                              Pull attestation bundle (OCI)
                                                       ↓
                                              verify.Run() (sig, schema, digest, chain)
                                                       ↓
                                              Allow / Deny (with reason)
```

## Prerequisites

- Kubernetes 1.25 or later.
- TLS certificates for the webhook endpoint (use [cert-manager](https://cert-manager.io/) or generate manually).
- An OCI registry containing published attestation bundles (see `llmsa publish`).
- The `llmsa` container image available to the cluster.

## Installation

### Using Helm

```bash
helm install llmsa-webhook deploy/helm/ \
  --namespace llmsa-system --create-namespace \
  --set registryPrefix=ghcr.io/your-org/attestations \
  --set tls.secretName=llmsa-webhook-tls
```

### Using Raw Manifests

```bash
kubectl apply -f deploy/webhook/namespace.yaml
kubectl apply -f deploy/webhook/serviceaccount.yaml
kubectl apply -f deploy/webhook/deployment.yaml
kubectl apply -f deploy/webhook/service.yaml
kubectl apply -f deploy/webhook/validatingwebhookconfiguration.yaml
```

Before applying the `ValidatingWebhookConfiguration`, inject the CA bundle:

```bash
CA_BUNDLE=$(kubectl get secret llmsa-webhook-tls -n llmsa-system \
  -o jsonpath='{.data.ca\.crt}')
sed -i "s/<CA_BUNDLE_BASE64>/$CA_BUNDLE/" deploy/webhook/validatingwebhookconfiguration.yaml
```

## Configuration

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `8443` | Webhook listen port |
| `--tls-cert` | | Path to TLS certificate file |
| `--tls-key` | | Path to TLS private key file |
| `--policy` | | Path to policy YAML file |
| `--schema-dir` | `schemas/v1` | Path to JSON schema directory |
| `--registry-prefix` | | OCI registry prefix for attestation bundle lookups |
| `--fail-open` | `false` | Allow pods through when verification encounters an error |
| `--cache-ttl-seconds` | `300` | Cache successful image verification results to reduce repeated OCI pulls |

## Namespace Opt-in

The webhook only intercepts resources in namespaces labelled with `llmsa-attestation: enabled`:

```bash
kubectl label namespace my-app llmsa-attestation=enabled
```

This prevents bootstrapping issues and allows gradual rollout across namespaces.

## Verification Flow

For each container image in the submitted resource:

1. **Extract image references** from the Pod spec (initContainers, containers, ephemeralContainers).
2. **Construct attestation OCI reference** using the registry prefix and image digest.
3. **Pull the attestation bundle** from the OCI registry.
4. **Run the four-stage verification** pipeline (signature, schema, digest, chain).
5. **Return allow or deny** with a descriptive message.

If any image fails verification, the entire resource is denied.

Successful verification decisions are cached per image-derived attestation reference for the configured TTL. This reduces admission latency for repeated deploys of the same digest and lowers registry load. Failed verifications are not cached.

## Fail-Open vs Fail-Closed

| Mode | Behaviour | Use Case |
|------|-----------|----------|
| Fail-closed (default) | Deny pods when verification errors occur | Production environments requiring strict enforcement |
| Fail-open (`--fail-open`) | Allow pods through on errors, log warnings | Gradual rollout, non-critical namespaces |

## Troubleshooting

**Webhook timeout errors**: Increase the webhook timeout in the `ValidatingWebhookConfiguration` (default 10s). Bundle pulls from remote registries may take longer.

**Certificate mismatch**: Ensure the CA bundle in the webhook configuration matches the TLS certificate used by the webhook server. Use cert-manager for automated rotation.

**Registry authentication**: The webhook uses the default Kubernetes image pull credentials. Configure `imagePullSecrets` on the webhook ServiceAccount or use Workload Identity.

**Missing attestations**: Ensure bundles are published with `llmsa publish` before deploying. Check the OCI reference format matches the registry prefix configuration.
