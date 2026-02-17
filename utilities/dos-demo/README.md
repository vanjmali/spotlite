# Denial-of-Service (DoS) Attack Demo

> **Educational / demonstration purposes only.**
> This script simulates a small-scale DoS attack against the Spotlite API
> gateway to show how rate limiting helps — and why it's not enough on its own.

## What it does

The demo runs in **three phases**:

### Phase 1 — Baseline

Sends 5 sequential `GET /api/users/healthz` requests and records normal response
times as a reference point (typically 5–50 ms).

### Phase 2 — Attack flood

Fires **5 waves of 20 concurrent requests** (100 total) to overwhelm the Traefik
rate-limiter's token bucket. Logs each wave showing how many requests succeed
(200) vs. get rate-limited (429), plus latency stats.

### Phase 3 — Collateral damage

Sends a **single "legitimate" request** immediately after the flood. Because the
attacker's burst already drained the shared token bucket, this request is
typically also blocked with `429 Too Many Requests` — proving that the DoS
attack denies service to everyone, not just the attacker.

## How the rate limiter works

Traefik's `rate-limit-global` middleware (defined in
`api-gateway/dynamic/middlewares.yml`) uses a **token bucket** algorithm:

| Parameter   | Value | Meaning                                     |
| ----------- | ----- | ------------------------------------------- |
| **burst**   | 10    | Max tokens in the bucket (initial capacity) |
| **average** | 2     | Tokens added per period                     |
| **period**  | 10s   | Refill interval                             |

The first 10 requests consume the burst. After that, only 2 new tokens appear
every 10 seconds. The 100-request flood drains the bucket almost instantly,
leaving nothing for legitimate users.

## Prerequisites

- The Spotlite Docker Compose stack must be running:
  ```bash
  docker compose up
  ```
- Node.js ≥ 18 (for native `fetch` and `AbortController` support).

## Usage

```bash
cd utilities/dos-demo
npm start
```

### Environment variables

| Variable        | Default                 | Description                     |
| --------------- | ----------------------- | ------------------------------- |
| `BASE_URL`      | `http://localhost:3000` | Base URL of the API gateway     |
| `WAVE_PAUSE_MS` | `200`                   | Pause (ms) between attack waves |

### Example

```bash
# Use defaults
npm start

# Faster flood (less pause between waves)
WAVE_PAUSE_MS=50 npm start

# Custom target
BASE_URL=http://192.168.1.50:3000 npm start
```

## Expected output

```
╔══════════════════════════════════════════════════════════════╗
║          ⚠  DENIAL-OF-SERVICE ATTACK DEMO  ⚠                ║
╚══════════════════════════════════════════════════════════════╝

Target:     http://localhost:3000/api/users/healthz
Baseline:   5 sequential requests
Attack:     5 waves × 20 concurrent = 100 requests
Collateral: 1 legitimate request after the flood
Total:      106 requests

 PHASE 1 — BASELINE   Measuring normal response times
   1.   OK      12ms  {"status":"ok"}
   2.   OK       8ms  {"status":"ok"}
  ...
  → Average baseline latency: 10ms

 PHASE 2 — ATTACK FLOOD   5 waves of 20 concurrent requests
  Wave 1/5  ██████████░░░░░░░░░░  10/10  avg   45ms  max   89ms
  Wave 2/5  ░░░░░░░░░░░░░░░░░░░░   0/20  avg    3ms  max    5ms
  ...

 PHASE 3 — COLLATERAL DAMAGE   One "legitimate" request right after the flood
   429      3ms  Too Many Requests

  ⚡ The legitimate request was ALSO blocked (429).

SUMMARY
  Total requests sent:     106
  ✓ Successful (200):       15
  ⚡ Rate-limited (429):    91

  RATE LIMITING ENGAGED   91/100 attack requests blocked (91%)
  COLLATERAL DAMAGE CONFIRMED   Legitimate user blocked by 429 after the flood.

KEY TAKEAWAYS
  1. Rate limiting is a double-edged sword.
  2. Response times degrade under load.
  3. Additional layers are needed for real protection:
     • Per-IP rate limiting (not just global)
     • WAF (Web Application Firewall)
     • CDN-level DDoS protection
     • Auto-scaling to absorb traffic spikes
     • Circuit breakers between services
```

## Key takeaways

- **Rate limiting works** — Traefik blocks the majority of the flood after the
  burst bucket is exhausted.
- **Collateral damage is real** — legitimate users sharing the same rate-limit
  bucket get denied too. The attack succeeds in its goal of denying service.
- **Defense in depth is essential** — a production system needs per-IP rate
  limiting, WAF rules, CDN-level DDoS protection, auto-scaling, and circuit
  breakers in addition to basic global rate limiting.
