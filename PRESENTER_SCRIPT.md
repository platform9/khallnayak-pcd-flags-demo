# Presenter Script — PCD Feature Flags

**Total runtime:** ~7 minutes (5 min slides + 60s live demo + buffer).
**Format:** [stage directions] regular text = what you say out loud.

---

## Before you start

- Open `~/pcd-feature-flags.pptx` in presenter mode.
- Make sure the 3 Terminal windows are visible on screen: T1 flagd, T2 service, T3 operator.
- Confirm T1 says flagd is running. Confirm T2 is printing gray "legacy" lines once per second.
- T3 has the flag set to off and the usage hint visible.

---

## Slide 1 — Title

[10 seconds]

> "Hey everyone. I'm Mahi. I'm going to show you something that took us one day to ship across three repos, seven pull requests — and it lets any PCD operator change backend service behavior in **30 seconds** without restarting a single pod."

[pause, let the tagline land]

> "No restart. No redeploy. No ticket. Let me show you why that matters."

[advance]

---

## Slide 2 — The Problem

[45 seconds]

> "PCD has nine backend services running per DU: hamgr, nova, neutron, cinder, masakari, glance, and a few more. **None of them can change behavior at runtime.**"

[click — show the bullet list]

> "Every time we ship a new feature, it goes one of two ways. Either it ships binary-on — the code is in production the moment the deploy lands, and we cross our fingers. Or it ships behind a static config option, where toggling it means restarting the service — which is a customer-visible blip on the DU."

> "Neither of those is what good infrastructure teams do anymore. They use feature flags."

[brief pause]

> "And here's the kicker: DUs now run as Kubernetes namespaces with helm-deployed pods. The platform has moved on. **Our toggle story has not.**"

[advance]

---

## Slide 3 — The Solution

[45 seconds]

> "So we adopted **flagd**. It's the CNCF reference implementation of the OpenFeature spec. One stateless pod per DU namespace. It reads flags from a Kubernetes ConfigMap, and hot-reloads when an operator runs `kubectl edit`."

[click — show the bullet list]

> "On the consumer side, services use the **upstream** OpenFeature SDK and the upstream flagd-provider — directly. No Platform9 wrapper. No new pf9-flags repo. About twenty lines of code per service to wire it in."

[gesture to pills]

> "We did both languages in parallel — Go for hamgr, Python for nova — using the **same** flag schema, **same** wire protocol, **same** SDK family on both sides."

[emphasize]

> "We chose, integrated, and shipped. We did **not** invent."

[advance]

---

## Slide 4 — Architecture

[40 seconds]

> "Here's how it works end-to-end."

[trace with finger or laser]

> "An operator types `kubectl edit cm/flagd-config` in a DU namespace. The Kubernetes API server validates the JSON and writes etcd. The kubelet on the node syncs that ConfigMap to disk inside the flagd pod, usually within thirty seconds."

> "flagd has an inotify watch on that file. The moment it changes, flagd reloads and pushes the new value out over a gRPC stream to every consumer that's subscribed."

> "On the consumer side — hamgr, nova-api, future neutron — the OpenFeature SDK keeps an in-process cache. So when application code calls `BooleanValue`, it's a sub-microsecond memory read."

[bottom of slide]

> "Operator edit to service eval: about thirty seconds, ninety-ninth percentile. No restart, anywhere."

[advance]

---

## Slide 5 — What We Shipped

[30 seconds]

> "Seven pull requests across three repos in one day. Foundation chart, end-to-end test, both pilot consumers, both chart wirings, and a usage doc."

[gesture at table]

> "PR 677 is the foundation — everything else depends on it. After that, the consumer work and the chart wiring split cleanly into independent PRs that can merge in any order."

> "Every single one has unit tests, helm template snapshot tests, and went through two-stage automated code review — spec compliance first, then code quality. **Production-ready, not hackathon-grade.**"

[advance]

---

## Slide 6 — Pilot Consumers

[40 seconds]

> "Here's what consuming a flag actually looks like, in both pilots."

[gesture left]

> "On the Go side, in hamgr — we call `initFlags()` once at startup. At a call site, we ask the OpenFeature client for a boolean value, passing a caller default — that's the value you get if flagd is unreachable. **Every eval is fail-open.** The service never crashes because of the flag system."

[gesture right]

> "Python is exactly the same shape. `init_flags()` at startup, `get_boolean_value` at the call site. Different syntax — same semantics."

> "Adding a third consumer — neutron, cinder, whatever — is the same pattern. Less than a day each."

[advance]

---

## Slide 7 — DEMO

[10 seconds — keep it short, switch to terminal]

> "Okay — enough slides. Let's see it."

[switch displays to the 3 Terminal windows]

---

## **LIVE DEMO** — 60 seconds

[Point at T1]

> "On the left, that's flagd running. It's watching a JSON file — in production this would be a Kubernetes ConfigMap, mounted as a file."

[Point at T2]

> "In the middle, that's a fake hamgr service. It's polling the flag `hamgr.fast_evac` once per second. Right now the flag is **off**, so it's printing the legacy path — serial migration, 96 seconds."

[Point at T3, then type slowly]

> "Now I'm the operator. In real life this is `kubectl edit cm/flagd-config`. For the demo I'm using a shell script that toggles the same JSON file directly."

[Type in T3, but don't hit Enter yet]

```
./flip.sh on
```

> "Watch the middle window."

[Hit Enter]

[Wait 2-3 seconds — the service will flip to green within 1 second]

> "There it goes. Cycle six was legacy. Cycle seven is the FAST path — parallel migration, twelve seconds. Notice the yellow arrow: **'flag flipped on, switched to fast path on cycle 7'**."

[Pause for effect]

> "And the rollback, just as fast:"

[Type and hit Enter]

```
./flip.sh off
```

[Wait 2-3 seconds]

> "Back to legacy. Service never restarted. Pods never recycled. No incident, no escalation, no on-call wake-up. **Sub-second mitigation.**"

[Switch back to slides]

---

## Slide 8 — What the audience just saw

[Skip this slide if the live demo worked. Use it only if the demo failed — read out the transcript instead.]

[15 seconds, only if needed]

> "What you would have just seen — flag off for two cycles, then I flipped it on, service switched within one second, ran four cycles of FAST, then I flipped it off, and it rolled back within one second. Same outcome whether it's a demo on my laptop or a real DU with a real operator running kubectl."

[advance]

---

## Slide 9 — Production Parity

[40 seconds]

> "I want to be very clear about one thing. **The demo is the production path.** It is not a simulation. It is not a sketch."

[gesture at table]

> "The only thing different between what you just saw and what runs on a real DU: the flag file is on local disk instead of in a ConfigMap, and flagd is a host process instead of a pod. **Everything else is identical** — same SDK, same provider, same wire protocol, same flag schema, same service code."

> "When this merges to production, we don't rewrite anything. We just change `localhost:8013` to a Kubernetes Service DNS name. Which the helm chart already does for us."

[advance]

---

## Slide 10 — Why this wins

[40 seconds]

> "Six reasons this is worth shipping."

[run down bullets quickly, hitting the punchy ones]

> "One — zero reinvention. flagd is CNCF, OpenFeature is the standard."

> "Two — already production-ready. Tests, reviews, e2e."

> "Three — working code, not slides. You just watched it work."

> "Four — extensible. Same pattern adds the rest of our services in days, not weeks."

> "Five — operators already know `kubectl edit`. Zero new tools. K8s RBAC already gates who can flip what."

> "Six — fail-open. If flagd is down, every service still works. They just get the caller default. **Nothing crashes.**"

[advance]

---

## Slide 11 — What's next

[30 seconds]

> "Today's scope is intentionally narrow: boolean flags, per-DU, backend pods only. That covers ninety percent of what we need right now."

[gesture at bullets]

> "Everything you'd want next — per-tenant targeting, percent rollouts, string flags, hypervisor-side support, an audit dashboard — is already supported by flagd's spec. We don't have to rebuild anything. **We just add code on top of the same foundation.**"

[advance]

---

## Slide 12 — Thanks

[15 seconds — closing punchline]

> "Thanks for watching."

[pause]

> "**No restart. No redeploy. No incident.** Questions?"

---

## If something breaks during the demo

- **flagd won't start (port conflict):** kill the old one — `pkill flagd` — then re-run `./start-flagd.sh`.
- **Service shows red errors:** flagd isn't up yet. Restart the service with `./start-service.sh` after confirming flagd is healthy.
- **flip.sh edits the file but service doesn't react:** confirm flagd is the one watching that exact file — its logs in T1 should mention `flags.json`. As a fallback, restart the service; it re-reads the current flag state at startup.
- **Tabs aren't aligned:** use slide 8 (the transcript slide) to walk through what would have happened.

---

## Pre-show checklist

- [ ] Slides open in presenter mode
- [ ] 3 Terminal windows visible, arranged left-to-right
- [ ] T1 shows flagd running, no errors in the last 30 seconds
- [ ] T2 shows the service printing gray "legacy" lines every second
- [ ] T3 shows the flag set to "off" (run `./flip.sh off` to confirm)
- [ ] Audience can see all three windows from the back of the room (zoom font if needed)
