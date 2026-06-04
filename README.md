# PCD Feature Flags — Hackathon Demo

> **Tagline:** "Flip a flag in 30 seconds. No restart. No redeploy. No ticket."

---

## The pitch (1 minute)

**Problem.** PCD has had no runtime feature-flag mechanism for backend services. The only flag system is `/opt/pf9/www/public/clarity/features.json` — driven by an ansible role, 12 years old, file-on-DU-host, only consumed by the Clarity UI. Backend services (hamgr, nova, neutron, cinder, masakari, glance) cannot gate behavior at runtime. Every new code path either ships binary-on (scary) or behind a static config that requires a service restart (slow).

Meanwhile, DUs now run as k8s namespaces with helm-deployed pods. The historical features.json shape is obsolete.

**Solution.** Adopt **flagd** — the CNCF OpenFeature reference engine. Stateless pod per DU namespace, ConfigMap-backed, hot-reloads on `kubectl edit`. Backend services consume the **upstream** OpenFeature SDK + flagd-provider directly. **No Platform9 wrapper. No new pf9-flags repo.** ~20 lines of code per service.

**Impact.**
- Operators flip a flag with `kubectl edit cm/flagd-config`. Effect visible service-side within ~30s.
- Every new feature can ship dark by default, then ramp at the operator's pace.
- Kill switches for risky code paths. Roll back without a redeploy.
- Both Go (pf9-hamgr) and Python (pf9-nova) services covered. Pattern extends trivially to neutron/cinder/glance/masakari.

**Status.** Implemented in 7 PRs on 2026-06-04. Pilot consumers (hamgr Go, nova Python) ship gated example features (verbose evac logs, verbose API request logs). Spec + plan + chart + e2e test + docs all in git.

| Repo | PR | What it ships |
|---|---|---|
| pf9-openstack-helm | [#677](https://github.com/platform9/pf9-openstack-helm/pull/677) | new `flagd/` helm chart (foundation) |
| pf9-openstack-helm | [#678](https://github.com/platform9/pf9-openstack-helm/pull/678) | kind-based e2e test |
| pf9-hamgr | [#41](https://github.com/platform9/pf9-hamgr/pull/41) | Go consumer + gated `hamgr.verbose_evac_logs` |
| pf9-openstack-helm | [#679](https://github.com/platform9/pf9-openstack-helm/pull/679) | hamgr chart env-var injection |
| pf9-nova | [#410](https://github.com/platform9/pf9-nova/pull/410) | Python consumer + gated `nova.verbose_api_request_log` |
| pf9-openstack-helm | [#680](https://github.com/platform9/pf9-openstack-helm/pull/680) | nova chart env-var injection |
| pf9-openstack-helm | [#681](https://github.com/platform9/pf9-openstack-helm/pull/681) | usage docs |

---

## The demo (60 seconds)

A simulated `hamgr` service loops every second, evaluates the flag `hamgr.fast_evac`, and prints which evacuation path it would take. We flip the flag from a third terminal. The audience sees the service switch behavior within ~1 second — no restart.

### Setup (one-time, already done on this laptop)

- `~/go/bin/flagd` — installed via `go install github.com/open-feature/flagd/flagd@latest`
- `/tmp/pcd-flags-demo/` — this directory

### Live demo — 3 terminals

Open three terminals side-by-side. Cwd in each: `/tmp/pcd-flags-demo`.

**Terminal 1 — flagd (the per-DU feature-flag engine)**

```bash
./start-flagd.sh
```

You'll see flagd's startup banner. It now watches `./flags.json` and serves gRPC on `:8013`.

**Terminal 2 — fake hamgr service**

```bash
./start-service.sh
```

Within a second you'll see lines like:

```
[19:51:25] cycle=5   legacy evac path: serial-migrating 8 VMs in 96s
[19:51:26] cycle=6   legacy evac path: serial-migrating 8 VMs in 96s
```

This is the flag-OFF behavior. The default.

**Terminal 3 — flip the flag (the operator's workflow)**

In the real product, the operator types `kubectl edit cm/flagd-config -n <du-ns>`. For the demo we simulate that with a one-liner that toggles the JSON file directly:

```bash
./flip.sh on
```

Within ~1 second, Terminal 2 lights up green:

```
[19:51:27] cycle=7   legacy evac path: serial-migrating 8 VMs in 96s
[19:51:28] cycle=8  🚀 FAST evac path: parallel-migrating 8 VMs in 12s
            ↑ flag flipped ON — switched to fast path on cycle 8
[19:51:29] cycle=9  🚀 FAST evac path: parallel-migrating 8 VMs in 12s
```

The service is now using the new code path.

**Roll back (for the punchline):**

```bash
./flip.sh off
```

Service flips back to legacy within ~1s. No restart. No redeploy. No incident.

---

## What's happening under the hood

```
flags.json (operator edits)
    │
    ▼  (file watch + inotify)
flagd  ── gRPC :8013 push ──>  service (OpenFeature SDK + flagd-provider)
                                      │
                                      ▼
                             client.BooleanValue("hamgr.fast_evac", false, ...)
```

In production: `flags.json` is mounted from a Kubernetes ConfigMap. kubelet syncs the ConfigMap to the pod's filesystem every ~30s; flagd inotify-detects the change and pushes the new value to every subscribed consumer over a gRPC stream. The consumer SDK keeps an in-process cache, so eval is sub-microsecond.

In the demo: we skip k8s and let flagd watch the JSON file directly. Same code path, same SDKs, same wire protocol — just no kubelet in the middle.

---

## Why this wins for a hackathon

1. **Zero Platform9 invention.** flagd is CNCF. OpenFeature is the upstream standard. We did not build a wrapper SDK, a new repo, or a custom DSL. We chose, integrated, and shipped — in 7 PRs across 3 repos in one day.
2. **Already in production-ready shape.** Each PR has unit tests, snapshot tests, and (for the chart) a kind-based e2e. Two reviewers (spec compliance + code quality) ran on every phase.
3. **Working code, not slides.** The demo runs on this laptop right now. Three terminals, three commands.
4. **Extensible.** Same pattern adds neutron/cinder/glance/masakari in <1 day each. Same pattern adds string/int/JSON flags. Same pattern adds per-tenant targeting in flagd's spec without changing a single consumer service.

---

## Files in this demo

| File | Purpose |
|---|---|
| `flags.json` | Flag definitions; flagd watches this file for changes |
| `service.go` | Fake hamgr that polls `hamgr.fast_evac` once per second |
| `start-flagd.sh` | Terminal 1: launches flagd on `:8013` |
| `start-service.sh` | Terminal 2: launches the fake service |
| `flip.sh` | Terminal 3: toggles the flag (`./flip.sh on` or `./flip.sh off`) |
| `go.mod`, `go.sum` | Pins OpenFeature go-sdk v1.11.0 + flagd-provider v0.2.4 |

---

## Production parity check

| Demo | Production |
|---|---|
| `flags.json` on local disk | ConfigMap `flagd-config` mounted at `/etc/flagd/flags.json` |
| flagd as a host process | flagd as a k8s Deployment (one pod per DU namespace) |
| Service runs `go run` against `localhost:8013` | hamgr/nova pod injects `FLAGD_HOST` / `FLAGD_PORT` via helm chart |
| `./flip.sh on` | `kubectl edit cm/flagd-config -n <du-ns>` |
| Same OpenFeature SDK | Same OpenFeature SDK |
| Same flagd-provider | Same flagd-provider |
| Same flag schema | Same flag schema |

The demo is the production path with k8s and the chart elided.
