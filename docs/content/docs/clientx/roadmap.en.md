---
title: 'roadmap'
linkTitle: 'roadmap'
description: 'clientx roadmap'
weight: 90
---

## clientx Roadmap (2026-03)

## Positioning

`clientx` is a protocol-oriented client layer, not a heavyweight RPC framework.

- Keep protocol-specific APIs explicit (`http` request/response, `tcp` stream, `udp` packet).
- Unify engineering conventions instead of call shapes (config validation/defaults, error model, policies, observability hooks).

## Snapshot (Updated)

Completed in current iteration:

- `New` constructors now validate and normalize configs with explicit defaults.
- `Close()` lifecycle capability is unified across `http`, `tcp`, and `udp` clients.
- `http.Execute` now requires `context.Context`; TCP TLS dial path also honors context cancellation.
- Capability interfaces are introduced in root package:
  - `clientx.Closer`
  - `clientx.Dialer`
  - `clientx.PacketListener`
- Policy pipeline foundation is introduced:
  - `clientx.Operation` / `OperationKind`
  - `clientx.Policy` / `PolicyFuncs`
  - `clientx.InvokeWithPolicies`
- `WithPolicies(...)` is wired into `http`, `tcp`, and `udp` clients.
- Built-in timeout-guard, retry/backoff, and concurrency-limit policies are implemented in `clientx`.
- Hook/policy panic isolation is enabled by default so single extension failures do not crash the client data path.
- `http.Config.Retry` is mapped to the unified policy pipeline, and `WithConcurrencyLimit(...)` / `WithTimeoutGuard(...)` are now available for HTTP/TCP/UDP.
- Engineering presets are introduced in `clientx/preset`: `NewEdgeHTTP`, `NewInternalRPC`, and `NewLowLatencyUDP` (with override options).

## Version Plan (Execution-Oriented)

- `v0.3.0-alpha.2` (completed)
  - Add built-in policy modules: timeout guard, retry/backoff, concurrency limit.
  - Add hook panic isolation (policy/hook failures must not crash client path by default).
- `v0.3.0-beta.1`
  - Standardize operation taxonomy and policy metadata conventions.
  - Add unified telemetry enrichment adapters for `observabilityx`.
- `v0.3.0-rc.1`
  - Publish end-to-end examples for service-to-service HTTP/TCP/UDP profiles.
  - Complete regression/perf baselines for policy overhead.

## Priority Suggestions

### P0 (Now)

- Finalize built-in policy set and default composition order.
- Harden preset default profiles and override rules in docs/examples.
- Expand tests for policy ordering, error-joining, and cancellation semantics.

### P1 (Next)

- Introduce context-aware hook contract and canonical operation attributes.
- Provide policy-level idempotency and retry classification helpers.
- Align docs and examples with capability-interface based usage.

### P2 (Later)

- Add optional transport extension points while keeping core lightweight.
- Add benchmark matrix for protocol + policy combinations.

## Non-Goals

- No one-size-fits-all abstraction that hides protocol semantics.
- No replacement of mature full-feature protocol SDKs.
- No forced dependency on one telemetry/retry implementation.





