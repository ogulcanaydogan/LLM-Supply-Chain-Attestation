# Quickstart

## 1) Bootstrap
```bash
make init
```

Optional for keyless Sigstore signing:
```bash
./scripts/install-cosign.sh
```

## 2) Generate attestations
```bash
make attest
```

## 3) Sign bundles
```bash
make sign
```

## 4) Verify and gate
```bash
make verify
make gate
```

## 5) Build Markdown report
```bash
make report
```

## 6) Run full demo
```bash
make demo
```
