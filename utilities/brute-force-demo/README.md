# Brute-Force Login Attack Demo

> **Educational / demonstration purposes only.**  
> This script simulates a dictionary-based brute-force attack against the
> Spotlite login endpoint to show how the Traefik API gateway's rate limiting
> mitigates such attacks.

## What it does

1. Sends ~25 `POST /api/users/login` requests in rapid succession, each with a
   different common/dictionary password and the target email address.
2. Logs every attempt with its HTTP status code and response time.
3. Demonstrates the transition from **401 Unauthorized** (wrong password) to
   **429 Too Many Requests** (rate-limited by Traefik) once the token bucket is
   exhausted.
4. Prints a summary showing how many requests were denied vs. rate-limited.

## How the rate limiter works

Traefik's `rate-limit-global` middleware (defined in
`api-gateway/dynamic/middlewares.yml`) uses a **token bucket** algorithm:

| Parameter   | Value | Meaning                                     |
| ----------- | ----- | ------------------------------------------- |
| **burst**   | 10    | Max tokens in the bucket (initial capacity) |
| **average** | 2     | Tokens added per period                     |
| **period**  | 10s   | Refill interval                             |

The first 10 requests consume the burst capacity. After that, only 2 new
requests are allowed every 10 seconds. Any request that arrives when the bucket
is empty gets a `429 Too Many Requests` response.

## Prerequisites

- The Spotlite Docker Compose stack must be running:
  ```bash
  docker compose up
  ```
- Node.js ≥ 18 (for native `fetch` support).

## Usage

```bash
cd utilities/brute-force-demo
npm start
```

### Environment variables

| Variable       | Default                      | Description                   |
| -------------- | ---------------------------- | ----------------------------- |
| `BASE_URL`     | `http://localhost:3000`      | Base URL of the API gateway   |
| `TARGET_EMAIL` | `teodora.nedic123@gmail.com` | Email address to attack       |
| `DELAY_MS`     | `80`                         | Milliseconds between attempts |

### Example

```bash
# Use defaults
npm start

# Custom base URL
BASE_URL=http://192.168.1.50 npm start

# Slower attack (200ms between attempts)
DELAY_MS=200 npm start
```

## Expected output

```
╔══════════════════════════════════════════════════════════════╗
║           ⚠  BRUTE-FORCE LOGIN ATTACK DEMO  ⚠               ║
╚══════════════════════════════════════════════════════════════╝

Target:    teodora.nedic123@gmail.com
Endpoint:  http://localhost/api/users/login
Passwords: 25 dictionary entries
Delay:     80ms between requests

 #   Password             Status   Latency  Response
───────────────────────────────────────────────────────────────────────────
 1.  pas*****              AUTH      12ms  {"error":"Invalid credentials."}
 2.  123***                AUTH       8ms  {"error":"Invalid credentials."}
 ...
11.  teo****               RATE      3ms  Too Many Requests
12.  Teo****               RATE      2ms  Too Many Requests
 ...

Summary

  ✗ Denied (401):          10
  ⚡ Rate-limited (429):    15
  ✓ Succeeded (200):       0

 RATE LIMITING DETECTED   Traefik started blocking at attempt #11.
 PASSWORD NOT FOUND   None of the 25 dictionary passwords matched.
```

## Key takeaways

- **Rate limiting works:** the API gateway blocks rapid-fire requests after the
  burst bucket is exhausted, making brute-force attacks impractical.
- **Strong passwords matter:** even without rate limiting, the dictionary attack
  fails because the account uses a strong, unique password (`Teodora123@`) that
  is not in any common wordlist.
- **Defense in depth:** rate limiting is one layer. In production you would also
  want account lockout policies, CAPTCHA, IP-based blocking, and monitoring /
  alerting.
