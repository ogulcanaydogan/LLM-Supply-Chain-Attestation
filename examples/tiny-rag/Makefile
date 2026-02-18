ROOT := ../..
ATTEST_DIR := $(ROOT)/.llmsa/attestations

.PHONY: bootstrap clean attest sign verify gate report tamper-prompt demo demo-bad-change demo-good-change

bootstrap:
	cd $(ROOT) && go run ./cmd/llmsa init

clean:
	rm -rf $(ATTEST_DIR) $(ROOT)/verify.json $(ROOT)/verify.md $(ROOT)/examples/tiny-rag/out/*

attest: bootstrap
	rm -rf $(ATTEST_DIR)
	mkdir -p $(ATTEST_DIR)
	cd $(ROOT) && go run ./cmd/llmsa attest create --type prompt_attestation --config examples/tiny-rag/configs/prompt.yaml --out .llmsa/attestations --determinism-check 2
	cd $(ROOT) && go run ./cmd/llmsa attest create --type corpus_attestation --config examples/tiny-rag/configs/corpus.yaml --out .llmsa/attestations --determinism-check 2
	cd $(ROOT) && go run ./cmd/llmsa attest create --type eval_attestation --config examples/tiny-rag/configs/eval.yaml --out .llmsa/attestations --determinism-check 2
	cd $(ROOT) && go run ./cmd/llmsa attest create --type route_attestation --config examples/tiny-rag/configs/route.yaml --out .llmsa/attestations --determinism-check 2
	cd $(ROOT) && go run ./cmd/llmsa attest create --type slo_attestation --config examples/tiny-rag/configs/slo.yaml --out .llmsa/attestations --determinism-check 2

sign: bootstrap
	cd $(ROOT) && for s in .llmsa/attestations/statement_*.json; do \
		case "$$s" in *.bundle.json) continue ;; esac; \
		b="$${s%.json}.bundle.json"; \
		go run ./cmd/llmsa sign --in "$$s" --provider sigstore --out "$$b"; \
	done

verify:
	cd $(ROOT) && go run ./cmd/llmsa verify --source local --attestations .llmsa/attestations --policy policy/examples/mvp-gates.yaml --format json --out verify.json

report:
	cd $(ROOT) && go run ./cmd/llmsa report --in verify.json --out verify.md

# gate can fail on repos without prior commit history; use HEAD~1 after first commit
gate:
	cd $(ROOT) && go run ./cmd/llmsa gate --policy policy/examples/mvp-gates.yaml --attestations .llmsa/attestations --git-ref HEAD~1

tamper-prompt:
	echo "tamper" >> $(ROOT)/examples/tiny-rag/app/system_prompt.txt

demo-bad-change: attest sign tamper-prompt
	cd $(ROOT) && go build -o .llmsa/llmsa ./cmd/llmsa
	cd $(ROOT) && set +e; .llmsa/llmsa verify --source local --attestations .llmsa/attestations --policy policy/examples/mvp-gates.yaml --format json --out verify.json; CODE=$$?; set -e; if [ $$CODE -ne 12 ]; then echo "expected exit 12, got $$CODE"; exit 1; fi

demo-good-change: attest sign verify report

# End-to-end target
demo: clean attest sign verify report
	@echo "demo completed"
