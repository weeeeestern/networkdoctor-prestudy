# AGENTS.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Repository reality check
- This repository is still documentation-heavy.
- Implementation sources currently include `bpf/networkdoctor.bpf.c` (kernel-side eBPF collector) and `cmd/networkdoctor-agent/main.go` (Go userspace loader / Prometheus exporter).
- Most content under `prestudy/` is study/reference markdown, not executable application code.

## Common commands
### Repository/documentation work
- Show top-level tracked files:
  - `git --no-pager ls-files`
- Inspect current changes:
  - `git --no-pager status`
  - `git --no-pager diff`

## eBPF collector and Go loader build (current code path)
`cmd/networkdoctor-agent/main.go` uses `go:generate` with `bpf/networkdoctor.bpf.c`.

1. Generate `vmlinux.h` from host BTF:
   - `bpftool btf dump file /sys/kernel/btf/vmlinux format c > bpf/vmlinux.h`
2. Generate Go BPF bindings:
   - `go generate ./...`
   - On non-Linux development hosts, use `go generate -tags linux ./...` so the Linux-only `main.go` package is included.
3. Build the userspace agent:
   - `go build ./cmd/networkdoctor-agent`

Notes:
- The checked-in BPF source path is `bpf/networkdoctor.bpf.c`.
- The old development copy used the typo `networkdoctro.bpf.c`; do not use that filename in this repository.
- This repo does not currently include a Makefile or unified build script.

## Lint/test status
- Go formatting is `gofmt -w cmd/networkdoctor-agent/main.go`.
- No repository-wide lint command is currently defined (no ESLint/ruff/golangci/etc. config checked in).
- No automated test suite is currently defined in this repository.
- There is no single-test command yet because no test runner is configured.

## High-level architecture
The project goal (from `README.md`) is a Kubernetes network incident diagnosis pipeline:

1. **Node-level signal collection** (DaemonSet + eBPF + kernel/network metrics)
2. **Prometheus ingestion** of node/service metrics
3. **Diagnosis backend** (anomaly detection + rule scoring + root-cause candidates)
4. **Presentation/deployment layer** (Grafana/Kiali/Incident UI + Helm/demo)

Current repository content maps to that target architecture like this:
- **Implemented artifact (partial)**: `bpf/networkdoctor.bpf.c`
  - Collects TCP, UDP, DNS, and scheduler/run-queue latency evidence via tracepoints/kprobe/TC hooks.
  - Stores counters/histograms/state in BPF maps and emits event records via ring buffer.
  - Exposes configurable thresholds through `nd_config` map defaults + overrides.
- **Design/reference artifacts**: `README.md`, `prestudy/**/*.md`
  - Document intended backend/rules/UI components and domain knowledge, but those components are not yet present as code in this repo.

When adding new code, keep a clear boundary between:
- `prestudy/` as reference/study material, and
- runtime implementation directories (`bpf/` for kernel eBPF code, `cmd/` for executable Go commands).

## Collaboration conventions already captured in this repo
Based on `prestudy/README.md` and `.github/pull_request_template.md`:
- Do not work directly on `main`; use topic branches and open PRs.
- Keep changes scoped to the intended topic.
- For markdown-heavy changes, verify formatting before PR.
