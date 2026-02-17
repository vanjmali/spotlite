#!/usr/bin/env node

/**
 * Denial-of-Service (DoS) Attack Demo
 * ─────────────────────────────────────
 * Simulates a DoS attack against the Spotlite API gateway in three phases:
 *
 *   Phase 1 — Baseline:  Measure normal response times with a few sequential
 *                         requests so we have a reference point.
 *   Phase 2 — Attack:    Fire concurrent request bursts (5 waves × 20 reqs)
 *                         to exhaust the Traefik rate-limiter bucket.
 *   Phase 3 — Collateral: Send a single "legitimate" request right after the
 *                          flood and show it is also blocked (429).
 *
 * The demo is capped at ~105 total requests — enough to drain the token bucket
 * and demonstrate degradation, but NOT enough to crash anything.
 *
 * Prerequisites:
 *   - The Spotlite Docker Compose stack must be running (`docker compose up`).
 *
 * Usage:
 *   node dos-demo.mjs
 *   BASE_URL=http://localhost:3000 npm start
 */

// ─── Configuration ───────────────────────────────────────────────────────────

const BASE_URL = "http://localhost:3000";
const TARGET_ENDPOINT = `${BASE_URL}/api/users/healthz`;

/** Number of sequential baseline requests in Phase 1. */
const BASELINE_COUNT = 5;

/** Number of concurrent requests per wave in Phase 2. */
const WAVE_SIZE = 20;

/** Number of attack waves in Phase 2. */
const WAVE_COUNT = 5;

/** Pause (ms) between waves — just enough for results to settle. */
const WAVE_PAUSE_MS = 200;

/** Per-request timeout (ms) — prevents hung connections. */
const REQUEST_TIMEOUT_MS = 5000;

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
  bgCyan: "\x1b[46m",
  white: "\x1b[37m",
};

// ─── Utilities ───────────────────────────────────────────────────────────────

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

function fmtMs(ms) {
  return `${String(ms).padStart(5)}ms`;
}

function statusColor(status) {
  if (status >= 200 && status < 300) return c.green;
  if (status === 429) return c.yellow;
  if (status >= 400) return c.red;
  return c.magenta;
}

function statusTag(status) {
  if (status === 0) return `${c.magenta} ERR  ${c.reset}`;
  if (status === 200) return `${c.green}  OK  ${c.reset}`;
  if (status === 429) return `${c.yellow}${c.bold} 429  ${c.reset}`;
  return `${statusColor(status)} ${status} ${c.reset}`;
}

/**
 * Send a single GET request with an AbortController timeout.
 * Returns { status, elapsed, body }.
 */
async function timedFetch(url) {
  const ac = new AbortController();
  const timer = setTimeout(() => ac.abort(), REQUEST_TIMEOUT_MS);

  const start = performance.now();
  let status, body;
  try {
    const res = await fetch(url, { signal: ac.signal });
    status = res.status;
    body = await res.text();
  } catch (err) {
    status = 0;
    body = err.name === "AbortError" ? "timeout" : err.message;
  } finally {
    clearTimeout(timer);
  }

  const elapsed = Math.round(performance.now() - start);
  return { status, elapsed, body: body.slice(0, 50) };
}

// ─── Graceful Ctrl+C ────────────────────────────────────────────────────────

let aborted = false;
process.on("SIGINT", () => {
  if (aborted) process.exit(1);
  aborted = true;
  console.log(
    `\n${c.yellow}Caught Ctrl+C — finishing current phase then exiting.${c.reset}\n`,
  );
});

// ─── Main ────────────────────────────────────────────────────────────────────

async function main() {
  // ── Banner ──
  console.log();
  console.log(
    `${c.bold}${c.red}╔══════════════════════════════════════════════════════════════╗${c.reset}`,
  );
  console.log(
    `${c.bold}${c.red}║          ⚠  DENIAL-OF-SERVICE ATTACK DEMO  ⚠                ║${c.reset}`,
  );
  console.log(
    `${c.bold}${c.red}╚══════════════════════════════════════════════════════════════╝${c.reset}`,
  );
  console.log();
  console.log(`${c.cyan}Target:${c.reset}     ${TARGET_ENDPOINT}`);
  console.log(
    `${c.cyan}Baseline:${c.reset}   ${BASELINE_COUNT} sequential requests`,
  );
  console.log(
    `${c.cyan}Attack:${c.reset}     ${WAVE_COUNT} waves × ${WAVE_SIZE} concurrent = ${WAVE_COUNT * WAVE_SIZE} requests`,
  );
  console.log(
    `${c.cyan}Collateral:${c.reset} 1 legitimate request after the flood`,
  );
  console.log(
    `${c.cyan}Total:${c.reset}      ${BASELINE_COUNT + WAVE_COUNT * WAVE_SIZE + 1} requests`,
  );
  console.log();
  console.log(
    `${c.dim}Traefik rate-limit config:  burst=10  average=2 tokens / 10s`,
  );
  console.log(`This demo floods the token bucket and then shows that even a`);
  console.log(`legitimate request afterwards is denied (429).${c.reset}`);

  // ── Collection for final summary ──
  const allResults = { baseline: [], attack: [], collateral: null };

  // ─────────────────────────────────────────────────────────────────────────
  //  PHASE 1 — Baseline
  // ─────────────────────────────────────────────────────────────────────────
  console.log();
  console.log(
    `${c.bgGreen}${c.bold}${c.white} PHASE 1 — BASELINE ${c.reset}  Measuring normal response times`,
  );
  console.log(`${c.dim}${"─".repeat(65)}${c.reset}`);

  for (let i = 0; i < BASELINE_COUNT && !aborted; i++) {
    const r = await timedFetch(TARGET_ENDPOINT);
    allResults.baseline.push(r);
    console.log(
      `  ${c.dim}${String(i + 1).padStart(2)}.${c.reset} ` +
        `${statusTag(r.status)}  ${c.dim}${fmtMs(r.elapsed)}${c.reset}  ` +
        `${c.dim}${r.body}${c.reset}`,
    );
    if (i < BASELINE_COUNT - 1) await sleep(300); // gentle spacing
  }

  const baselineAvg = Math.round(
    allResults.baseline.reduce((s, r) => s + r.elapsed, 0) /
      allResults.baseline.length,
  );
  console.log();
  console.log(
    `  ${c.green}→ Average baseline latency: ${baselineAvg}ms${c.reset}`,
  );

  if (aborted) return printSummary(allResults, baselineAvg);

  // ─────────────────────────────────────────────────────────────────────────
  //  PHASE 2 — Attack waves
  // ─────────────────────────────────────────────────────────────────────────
  console.log();
  console.log(
    `${c.bgRed}${c.bold}${c.white} PHASE 2 — ATTACK FLOOD ${c.reset}  ${WAVE_COUNT} waves of ${WAVE_SIZE} concurrent requests`,
  );
  console.log(`${c.dim}${"─".repeat(65)}${c.reset}`);

  for (let wave = 0; wave < WAVE_COUNT && !aborted; wave++) {
    const waveLabel = `Wave ${wave + 1}/${WAVE_COUNT}`;

    // Fire all requests in this wave concurrently
    const promises = Array.from({ length: WAVE_SIZE }, () =>
      timedFetch(TARGET_ENDPOINT),
    );
    const results = await Promise.all(promises);

    allResults.attack.push(...results);

    // Tally this wave
    const ok = results.filter((r) => r.status === 200).length;
    const limited = results.filter((r) => r.status === 429).length;
    const errs = results.filter(
      (r) => r.status !== 200 && r.status !== 429,
    ).length;
    const avgMs = Math.round(
      results.reduce((s, r) => s + r.elapsed, 0) / results.length,
    );
    const maxMs = Math.max(...results.map((r) => r.elapsed));

    const bar200 = "█".repeat(ok);
    const bar429 = "░".repeat(limited);
    const barErr = "?".repeat(errs);

    console.log(
      `  ${c.bold}${waveLabel}${c.reset}  ` +
        `${c.green}${bar200}${c.reset}${c.yellow}${bar429}${c.reset}${c.magenta}${barErr}${c.reset}  ` +
        `${c.green}${ok}${c.reset}/${c.yellow}${limited}${c.reset}` +
        `${errs ? `/${c.magenta}${errs}${c.reset}` : ""}  ` +
        `avg ${c.dim}${fmtMs(avgMs)}${c.reset}  ` +
        `max ${c.dim}${fmtMs(maxMs)}${c.reset}`,
    );

    if (wave < WAVE_COUNT - 1) await sleep(WAVE_PAUSE_MS);
  }

  if (aborted) return printSummary(allResults, baselineAvg);

  // ─────────────────────────────────────────────────────────────────────────
  //  PHASE 3 — Collateral damage
  // ─────────────────────────────────────────────────────────────────────────
  console.log();
  console.log(
    `${c.bgYellow}${c.bold} PHASE 3 — COLLATERAL DAMAGE ${c.reset}  ` +
      `One "legitimate" request right after the flood`,
  );
  console.log(`${c.dim}${"─".repeat(65)}${c.reset}`);

  const legit = await timedFetch(TARGET_ENDPOINT);
  allResults.collateral = legit;

  console.log(
    `  ${statusTag(legit.status)}  ${c.dim}${fmtMs(legit.elapsed)}${c.reset}  ${c.dim}${legit.body}${c.reset}`,
  );
  console.log();

  if (legit.status === 429) {
    console.log(
      `  ${c.yellow}${c.bold}⚡ The legitimate request was ALSO blocked (429).${c.reset}`,
    );
    console.log(
      `  ${c.dim}The attacker's flood drained the shared rate-limit bucket,`,
    );
    console.log(
      `  so real users are denied service too — this IS the denial of service.${c.reset}`,
    );
  } else if (legit.status === 200) {
    console.log(
      `  ${c.green}The legitimate request got through (a token was refilled).${c.reset}`,
    );
    console.log(
      `  ${c.dim}Try reducing WAVE_PAUSE_MS to make the flood more aggressive.${c.reset}`,
    );
  }

  printSummary(allResults, baselineAvg);
}

// ─── Summary ─────────────────────────────────────────────────────────────────

function printSummary(results, baselineAvg) {
  const all = [...results.baseline, ...results.attack];
  if (results.collateral) all.push(results.collateral);

  const total = all.length;
  const ok = all.filter((r) => r.status === 200).length;
  const limited = all.filter((r) => r.status === 429).length;
  const errs = all.filter((r) => r.status !== 200 && r.status !== 429).length;

  const attackOk = results.attack.filter((r) => r.status === 200).length;
  const attackLimited = results.attack.filter((r) => r.status === 429).length;
  const attackAvg = results.attack.length
    ? Math.round(
        results.attack.reduce((s, r) => s + r.elapsed, 0) /
          results.attack.length,
      )
    : 0;

  console.log();
  console.log(`${c.dim}${"─".repeat(65)}${c.reset}`);
  console.log(`${c.bold}SUMMARY${c.reset}`);
  console.log();
  console.log(`  Total requests sent:     ${total}`);
  console.log(`  ${c.green}✓${c.reset} Successful (200):       ${ok}`);
  console.log(`  ${c.yellow}⚡${c.reset} Rate-limited (429):    ${limited}`);
  if (errs > 0) {
    console.log(`  ${c.magenta}⚠${c.reset} Errors:                ${errs}`);
  }
  console.log();
  console.log(`  ${c.cyan}Baseline avg latency:${c.reset}    ${baselineAvg}ms`);
  console.log(`  ${c.cyan}Attack avg latency:${c.reset}      ${attackAvg}ms`);
  console.log();

  if (attackLimited > 0) {
    const pct = Math.round((attackLimited / results.attack.length) * 100);
    console.log(
      `  ${c.bgYellow}${c.bold} RATE LIMITING ENGAGED ${c.reset}  ` +
        `${attackLimited}/${results.attack.length} attack requests blocked (${pct}%)`,
    );
  }

  if (results.collateral?.status === 429) {
    console.log(
      `  ${c.bgRed}${c.bold} COLLATERAL DAMAGE CONFIRMED ${c.reset}  ` +
        `Legitimate user blocked by 429 after the flood.`,
    );
  }

  // ── Takeaways ──
  console.log();
  console.log(`${c.bold}KEY TAKEAWAYS${c.reset}`);
  console.log();
  console.log(`  1. ${c.bold}Rate limiting is a double-edged sword.${c.reset}`);
  console.log(`     It blocks the flood, but also blocks legitimate users who`);
  console.log(`     share the same rate-limit bucket.`);
  console.log();
  console.log(`  2. ${c.bold}Response times degrade under load.${c.reset}`);
  console.log(
    `     Baseline: ~${baselineAvg}ms → Attack avg: ~${attackAvg}ms.`,
  );
  if (attackAvg > baselineAvg * 1.5) {
    console.log(
      `     That's a ${c.red}${Math.round(attackAvg / baselineAvg)}× slowdown${c.reset} for requests that do get through.`,
    );
  }
  console.log();
  console.log(
    `  3. ${c.bold}Additional layers are needed for real protection:${c.reset}`,
  );
  console.log(`     • Per-IP rate limiting (not just global)`);
  console.log(`     • WAF (Web Application Firewall)`);
  console.log(
    `     • CDN-level DDoS protection (Cloudflare, AWS Shield, etc.)`,
  );
  console.log(`     • Auto-scaling to absorb traffic spikes`);
  console.log(`     • Circuit breakers between services`);
  console.log();
}

// ─── Run ─────────────────────────────────────────────────────────────────────

main().catch((err) => {
  console.error(`${c.red}Fatal:${c.reset}`, err);
  process.exit(1);
});
