import fs from 'node:fs';
import path from 'node:path';
import { execSync } from 'node:child_process';

const ROOT_DIR = process.cwd();
const ENV_PATH = path.join(ROOT_DIR, '.env');
const envFile = fs.existsSync(ENV_PATH) ? fs.readFileSync(ENV_PATH, 'utf8') : '';
const envFromFile = parseDotEnv(envFile);

const cfg = {
  apiBaseUrl:
    process.env.API_BASE_URL ||
    process.env.APP_API_BASE_URL ||
    envFromFile.APP_API_BASE_URL ||
    'https://localhost:4443/api',
  mailhogUrl: process.env.MAILHOG_URL || 'http://localhost:8025',
  adminEmail: process.env.SEED_ADMIN_EMAIL || 'admin@example.com',
  adminPassword: process.env.SEED_ADMIN_PASSWORD || 'Example123!',
  adminUsername: process.env.SEED_ADMIN_USERNAME || 'admin',
  firstName: process.env.SEED_ADMIN_FIRST_NAME || 'Admin',
  lastName: process.env.SEED_ADMIN_LAST_NAME || 'User',
  dockerComposeCmd: process.env.DOCKER_COMPOSE_CMD || 'docker compose',
  userMongoRootUser:
    process.env.SRV_USER_MONGO_ROOT_USERNAME || envFromFile.SRV_USER_MONGO_ROOT_USERNAME || '',
  userMongoRootPass:
    process.env.SRV_USER_MONGO_ROOT_PASSWORD || envFromFile.SRV_USER_MONGO_ROOT_PASSWORD || '',
  userMongoDb: process.env.SRV_USER_MONGO_DATABASE || envFromFile.SRV_USER_MONGO_DATABASE || '',
  userMongoService: process.env.SRV_USER_MONGO_SERVICE || 'user-mongodb',
  debug: process.env.SEED_DEBUG !== '0',
  max429Retries: Number(process.env.SEED_MAX_429_RETRIES || 6),
  base429BackoffMs: Number(process.env.SEED_BASE_429_BACKOFF_MS || 800),
};

const SONG_SPECS = [
  {
    title: 'The Builder',
    description:
      'Comedic, free-wheeling fun! Attribution Code\n"The Builder" Kevin MacLeod (incompetech.com)\nLicensed under Creative Commons: By Attribution 4.0 License\nhttp://creativecommons.org/licenses/by/4.0/',
    url: 'https://incompetech.com/music/royalty-free/mp3-royaltyfree/The%20Builder.mp3',
  },
  {
    title: 'Monkeys Spinning Monkeys',
    description:
      'Loopable happy light fluffy piece with bright flutes and a bunch of pizzicato strings.\nAttribution Code\n"Monkeys Spinning Monkeys" Kevin MacLeod (incompetech.com)\nLicensed under Creative Commons: By Attribution 4.0 License\nhttp://creativecommons.org/licenses/by/4.0/',
    url: 'https://incompetech.com/music/royalty-free/mp3-royaltyfree/Monkeys%20Spinning%20Monkeys.mp3',
  },
  {
    title: 'Merry Go',
    description:
      'Comedic and playful, this rag-time ditty has a strong melody, and is heavy in the bass chords. The second minute features flighty finger-work, as if an energetic bee is flying up and down the scales. The last thirty seconds is a refrain of the introduction, and the piece ends with an invigorating flourish.\nInstruments: Piano\nAttribution Code\n"Merry Go" Kevin MacLeod (incompetech.com)\nLicensed under Creative Commons: By Attribution 4.0 License\nhttp://creativecommons.org/licenses/by/4.0/',
    url: 'https://incompetech.com/music/royalty-free/mp3-royaltyfree/Merry%20Go.mp3',
  },
  {
    title: 'Spazzmatica Polka',
    description:
      'Boisterous and nearly obnoxious, this piece will lodge itself in your brain and make you think you?re trapped in an arcade. The polka rhythm is quick and the melody is spastic. Right into the second minute the rhythm drops out and introduces a crazed, comedic melody that continues throughout until the abrupt end.\nInstruments: Synths\nAttribution Code\n"Spazzmatica Polka" Kevin MacLeod (incompetech.com)\nLicensed under Creative Commons: By Attribution 4.0 License\nhttp://creativecommons.org/licenses/by/4.0/',
    url: 'https://incompetech.com/music/royalty-free/mp3-royaltyfree/Spazzmatica%20Polka.mp3',
  },
];

const FIXED_CONTENT = {
  genre: 'Soundtrack',
  artist: 'Kevin MacLeod',
  album: 'Comedic (Film Scoring Moods)',
  artistDescription:
    'Composer known for royalty-free music used across film, games, streaming, and educational projects.',
};

const EXTRA_GENRES = ['Electronic', 'Ambient', 'Indie Rock', 'Lo-fi', 'Synthwave', 'Jazz Hop'];
const EXTRA_ARTISTS = [
  'Neon Harbor',
  'Velvet Arrays',
  'Copper Skyline',
  'Glass Satellites',
  'Midnight Transit',
];
const AUDIO_BUFFER_CACHE = new Map();

async function main() {
  maybeAllowInsecureTls(cfg.apiBaseUrl);
  validateConfig();

  log(`API base: ${cfg.apiBaseUrl}`);
  debug('Effective config:', {
    apiBaseUrl: cfg.apiBaseUrl,
    mailhogUrl: cfg.mailhogUrl,
    adminEmail: cfg.adminEmail,
    userMongoService: cfg.userMongoService,
    userMongoDb: cfg.userMongoDb,
  });
  await ensureAdminUser();
  const adminToken = await loginAsAdminAndGetToken();

  const soundtrackGenre = await ensureGenre(adminToken, FIXED_CONTENT.genre);
  const kevinArtist = await ensureArtist(
    adminToken,
    FIXED_CONTENT.artist,
    [soundtrackGenre.id],
    FIXED_CONTENT.artistDescription
  );
  const comedicAlbum = await ensureAlbum(
    adminToken,
    FIXED_CONTENT.album,
    randomReleaseDate(),
    [soundtrackGenre.id],
    [kevinArtist.id]
  );

  for (const songSpec of SONG_SPECS) {
    const audioBuffer = await getOrDownloadAudioBuffer(songSpec);
    await ensureSongWithAudio(
      adminToken,
      {
        title: songSpec.title,
        albumId: comedicAlbum.id,
        genreIds: [soundtrackGenre.id],
        artistIds: [kevinArtist.id],
      },
      `${toSafeFilename(songSpec.title)}.mp3`,
      audioBuffer
    );
    log(`Ensured song "${songSpec.title}" in album "${FIXED_CONTENT.album}"`);
  }

  // Add extra entities so the app has browseable data.
  const extraGenreIds = [];
  for (const genreName of EXTRA_GENRES) {
    const g = await ensureGenre(adminToken, genreName);
    extraGenreIds.push(g.id);
  }

  for (let i = 0; i < EXTRA_ARTISTS.length; i += 1) {
    const artistName = EXTRA_ARTISTS[i];
    const genreId = extraGenreIds[i % extraGenreIds.length];
    const artist = await ensureArtist(
      adminToken,
      artistName,
      [genreId],
      `${artistName} is a generated demo artist for local development and testing.`
    );
    const album = await ensureAlbum(
      adminToken,
      `${artistName} - Vol. ${i + 1}`,
      randomReleaseDate(),
      [genreId],
      [artist.id]
    );

    // Keep generated albums non-empty by attaching one seeded audio track.
    const chosen = pickSeedSongSpec(`${artistName}:${album.id}`);
    const audioBuffer = await getOrDownloadAudioBuffer(chosen);
    await ensureSongWithAudio(
      adminToken,
      {
        title: `${chosen.title} (${artistName} Demo)`,
        albumId: album.id,
        genreIds: [genreId],
        artistIds: [artist.id],
      },
      `${toSafeFilename(`${chosen.title}-${artistName}-demo`)}.mp3`,
      audioBuffer
    );
    log(`Ensured demo song in album "${album.title}"`);
  }

  log('Complete.');
  log(`Admin login: ${cfg.adminEmail} / ${cfg.adminPassword}`);
}

async function ensureAdminUser() {
  await registerUserIfMissing();
  promoteUserToAdminAndActivate(cfg.adminEmail);
}

async function registerUserIfMissing() {
  const payload = {
    email: cfg.adminEmail,
    password: cfg.adminPassword,
    username: cfg.adminUsername,
    first_name: cfg.firstName,
    last_name: cfg.lastName,
  };

  const res = await fetchJson(`${cfg.apiBaseUrl}/users/register`, {
    method: 'POST',
    body: JSON.stringify(payload),
    headers: {
      'Content-Type': 'application/json',
    },
    allowError: true,
  });

  if (res.ok) {
    log('Registered admin user.');
    return;
  }

  if (res.status === 409) {
    log('Admin user already exists.');
    return;
  }

  throw error(`Failed registering admin user: ${res.status} ${res.text}`);
}

function promoteUserToAdminAndActivate(email) {
  const escapedEmail = email.replace(/\\/g, '\\\\').replace(/"/g, '\\"');
  runMongoEval(`
    db.users.updateOne(
      { email: "${escapedEmail}" },
      {
        $set: {
          role: "ADMIN",
          account_status: "ACTIVE"
        }
      }
    )
  `);
  log('Ensured admin role + active status in MongoDB.');
}

async function loginAsAdminAndGetToken() {
  const loginRes = await fetchJson(`${cfg.apiBaseUrl}/users/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      email: cfg.adminEmail,
      password: cfg.adminPassword,
    }),
  });

  if (!loginRes.ok) {
    throw error(`Login failed: ${loginRes.status} ${loginRes.text}`);
  }
  debug('Login requested, waiting for OTP...');

  const otp = await pollOtp(cfg.adminEmail);
  const verifyRes = await fetchJson(`${cfg.apiBaseUrl}/users/login/verify-otp`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      email: cfg.adminEmail,
      code: otp,
    }),
  });

  if (!verifyRes.ok || !verifyRes.json?.access_token) {
    throw error(`OTP verify failed: ${verifyRes.status} ${verifyRes.text}`);
  }

  log('Logged in as admin.');
  return verifyRes.json.access_token;
}

async function pollOtp(email, timeoutMs = 60000) {
  const started = Date.now();
  while (Date.now() - started < timeoutMs) {
    const res = await fetchJson(
      `${cfg.mailhogUrl}/api/v2/search?kind=to&query=${encodeURIComponent(email)}`,
      { allowError: true }
    );

    if (res.ok && res.json?.items?.length) {
      debug(`MailHog found ${res.json.items.length} messages for ${email}`);
      const body =
        res.json.items[0]?.Content?.Body || res.json.items[0]?.Raw?.Data || res.json.items[0]?.MIME?.Body || '';

      const otpFromHtml = body.match(/otp-code[^>]*>(\d{6})</);
      if (otpFromHtml) {
        return otpFromHtml[1];
      }

      const otpRaw = body.match(/(?:^|[^\d#])(\d{6})(?!\d)/);
      if (otpRaw) {
        return otpRaw[1];
      }
    }

    await sleep(1500);
  }

  throw error('Timed out waiting for OTP email in MailHog.');
}

async function ensureGenre(token, genreName) {
  log(`Ensuring genre "${genreName}"...`);
  const existing = await findGenreByName(token, genreName);
  if (existing) {
    debug(`Genre exists: ${existing.id}`);
    return existing;
  }

  const create = await authedJson(`${cfg.apiBaseUrl}/content/genres`, token, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name: genreName }),
    allowError: true,
  });
  if (!create.ok && create.status !== 409) {
    throw error(`Failed creating genre "${genreName}": ${create.status} ${create.text}`);
  }

  const created = await findGenreByName(token, genreName);
  if (!created) {
    throw error(`Could not resolve genre after create: ${genreName}`);
  }
  return created;
}

async function ensureArtist(token, name, genreIds, description) {
  log(`Ensuring artist "${name}"...`);
  const existing = await findArtistByName(token, name);
  if (existing) {
    debug(`Artist exists: ${existing.id}`);
    return existing;
  }

  const create = await authedJson(`${cfg.apiBaseUrl}/content/artists`, token, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      name,
      genre_ids: genreIds,
      description,
    }),
  });
  if (!create.ok) {
    throw error(`Failed creating artist "${name}": ${create.status} ${create.text}`);
  }

  const created = await findArtistByName(token, name);
  if (!created) {
    throw error(`Could not resolve artist after create: ${name}`);
  }
  return created;
}

async function ensureAlbum(token, title, releaseDate, genreIds, artistIds) {
  log(`Ensuring album "${title}"...`);
  const existing = await findAlbumByTitle(token, title);
  if (existing) {
    debug(`Album exists: ${existing.id}`);
    return existing;
  }

  const create = await authedJson(`${cfg.apiBaseUrl}/content/albums`, token, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      title,
      release_date: releaseDate,
      genre_ids: genreIds,
      artist_ids: artistIds,
    }),
  });
  if (!create.ok) {
    throw error(`Failed creating album "${title}": ${create.status} ${create.text}`);
  }

  const created = await retryFind(
    async () => findAlbumByTitle(token, title),
    6,
    500,
    `album "${title}"`
  );
  if (!created) {
    throw error(`Could not resolve album after create: ${title}`);
  }
  return created;
}

async function ensureSongWithAudio(token, songMeta, fileName, audioBuffer) {
  log(`Ensuring song "${songMeta.title}"...`);
  const albumSongsRes = await authedJson(
    `${cfg.apiBaseUrl}/content/albums/${songMeta.albumId}/songs`,
    token
  );
  if (!albumSongsRes.ok) {
    throw error(`Failed loading album songs: ${albumSongsRes.status} ${albumSongsRes.text}`);
  }

  const existingSong = (albumSongsRes.json || []).find(
    (song) => normalize(song.title) === normalize(songMeta.title)
  );

  if (!existingSong) {
    debug(`Song "${songMeta.title}" missing in album, creating with audio...`);
    const fd = new FormData();
    fd.append(
      'meta',
      JSON.stringify({
        title: songMeta.title,
        album_id: songMeta.albumId,
        genre_ids: songMeta.genreIds,
        artist_ids: songMeta.artistIds,
      })
    );
    fd.append('file', new Blob([audioBuffer], { type: 'audio/mpeg' }), fileName);

    const create = await authedJson(`${cfg.apiBaseUrl}/content/songs`, token, {
      method: 'POST',
      body: fd,
    });

    if (!create.ok) {
      throw error(`Failed creating song "${songMeta.title}": ${create.status} ${create.text}`);
    }
    return;
  }

  debug(`Song exists (${existingSong.id}), uploading/replacing audio...`);
  const fd = new FormData();
  fd.append('file', new Blob([audioBuffer], { type: 'audio/mpeg' }), fileName);
  const upload = await authedJson(`${cfg.apiBaseUrl}/content/songs/${existingSong.id}/audio`, token, {
    method: 'POST',
    body: fd,
  });
  if (!upload.ok) {
    throw error(
      `[seed] Failed uploading audio for "${songMeta.title}" (${existingSong.id}): ${upload.status} ${upload.text}`
    );
  }
}

async function findGenreByName(token, name) {
  const res = await authedJson(
    `${cfg.apiBaseUrl}/content/genres?name=${encodeURIComponent(name)}&page=1&size=100`,
    token
  );
  if (!res.ok) {
    throw error(`Failed listing genres: ${res.status} ${res.text}`);
  }
  const items = res.json?.items || [];
  let match = items.find((g) => normalize(g.name) === normalize(name));
  if (match) {
    return match;
  }

  const scanned = await listAllPaginated(`${cfg.apiBaseUrl}/content/genres`, token, {
    size: 200,
  });
  match = scanned.find((g) => normalize(g.name) === normalize(name));
  return match || null;
}

async function findArtistByName(token, name) {
  const res = await authedJson(
    `${cfg.apiBaseUrl}/content/artists?name=${encodeURIComponent(name)}&page=1&size=100`,
    token
  );
  if (!res.ok) {
    throw error(`Failed listing artists: ${res.status} ${res.text}`);
  }
  const items = res.json?.items || [];
  let match = items.find((a) => normalize(a.name) === normalize(name));
  if (match) {
    return match;
  }

  const scanned = await listAllPaginated(`${cfg.apiBaseUrl}/content/artists`, token, {
    size: 200,
  });
  match = scanned.find((a) => normalize(a.name) === normalize(name));
  return match || null;
}

async function findAlbumByTitle(token, title) {
  const res = await authedJson(
    `${cfg.apiBaseUrl}/content/albums?title=${encodeURIComponent(title)}&page=1&size=100`,
    token
  );
  if (!res.ok) {
    throw error(`Failed listing albums: ${res.status} ${res.text}`);
  }
  const items = res.json?.items || [];
  let match = items.find((a) => normalize(a.title) === normalize(title));
  if (match) {
    return match;
  }

  const scanned = await listAllPaginated(`${cfg.apiBaseUrl}/content/albums`, token, {
    size: 200,
  });
  match = scanned.find((a) => normalize(a.title) === normalize(title));
  return match || null;
}

async function authedJson(url, token, options = {}) {
  const headers = {
    ...(options.headers || {}),
    Authorization: `Bearer ${token}`,
  };
  return fetchJson(url, { ...options, headers });
}

async function fetchJson(url, options = {}) {
  const method = options.method || 'GET';
  const shouldRetry429 = options.retry429 !== false;
  let last = null;

  for (let attempt = 0; attempt <= cfg.max429Retries; attempt += 1) {
    debug(`HTTP ${method} ${url}${attempt > 0 ? ` (retry ${attempt}/${cfg.max429Retries})` : ''}`);
    const res = await fetch(url, options);
    const text = await res.text();
    let json = null;
    if (text) {
      try {
        json = JSON.parse(text);
      } catch {
        json = null;
      }
    }

    last = { ok: res.ok, status: res.status, json, text };
    debug(`HTTP ${res.status} ${url}`);

    if (res.status === 429 && shouldRetry429 && attempt < cfg.max429Retries) {
      const delay = get429DelayMs(attempt);
      log(`Rate limited (${method} ${url}). Retrying in ${delay}ms...`);
      await sleep(delay);
      continue;
    }

    if (!res.ok && !options.allowError) {
      debug(`HTTP FAIL ${res.status} ${url} :: ${text}`);
      throw error(`${res.status} ${res.statusText} - ${text}`);
    }

    return last;
  }

  if (!last) {
    throw error(`No HTTP response for ${method} ${url}`);
  }
  if (!last.ok && !options.allowError) {
    throw error(`${last.status} response after retries: ${last.text}`);
  }
  return last;
}

async function downloadBuffer(url) {
  const res = await fetch(url);
  if (!res.ok) {
    throw error(`Failed downloading file: ${res.status} ${url}`);
  }
  const arr = await res.arrayBuffer();
  return Buffer.from(arr);
}

async function getOrDownloadAudioBuffer(songSpec) {
  const cached = AUDIO_BUFFER_CACHE.get(songSpec.url);
  if (cached) {
    return cached;
  }

  log(`Downloading audio: ${songSpec.title}`);
  const audioBuffer = await downloadBuffer(songSpec.url);
  AUDIO_BUFFER_CACHE.set(songSpec.url, audioBuffer);
  return audioBuffer;
}

function pickSeedSongSpec(seed) {
  let hash = 0;
  for (let i = 0; i < seed.length; i += 1) {
    hash = (hash * 31 + seed.charCodeAt(i)) >>> 0;
  }
  return SONG_SPECS[hash % SONG_SPECS.length];
}

function runMongoEval(jsCode) {
  const cmd = [
    cfg.dockerComposeCmd,
    'exec -T',
    cfg.userMongoService,
    'mongosh',
    '--quiet',
    `--username "${cfg.userMongoRootUser}"`,
    `--password "${cfg.userMongoRootPass}"`,
    '--authenticationDatabase admin',
    cfg.userMongoDb,
    `--eval '${jsCode.replace(/\n/g, ' ').replace(/'/g, "\\'")}'`,
  ].join(' ');

  debug(`Mongo update command: ${cfg.dockerComposeCmd} exec -T ${cfg.userMongoService} mongosh ...`);
  execSync(cmd, { stdio: 'pipe', cwd: ROOT_DIR });
}

function validateConfig() {
  if (!cfg.userMongoRootUser || !cfg.userMongoRootPass || !cfg.userMongoDb) {
    throw error(
      '[seed] Missing user-service Mongo root credentials. Set SRV_USER_MONGO_ROOT_USERNAME, SRV_USER_MONGO_ROOT_PASSWORD, SRV_USER_MONGO_DATABASE in .env.'
    );
  }
}

function parseDotEnv(content) {
  const out = {};
  for (const rawLine of content.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line || line.startsWith('#')) {
      continue;
    }
    const idx = line.indexOf('=');
    if (idx <= 0) {
      continue;
    }
    const key = line.slice(0, idx).trim();
    let value = line.slice(idx + 1).trim();
    if (
      (value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))
    ) {
      value = value.slice(1, -1);
    }
    out[key] = value;
  }
  return out;
}

function normalize(v) {
  return String(v || '').trim().toLowerCase();
}

function toSafeFilename(name) {
  return name.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)/g, '');
}

function randomReleaseDate() {
  const year = 2016 + Math.floor(Math.random() * 10);
  const month = `${1 + Math.floor(Math.random() * 12)}`.padStart(2, '0');
  const day = `${1 + Math.floor(Math.random() * 28)}`.padStart(2, '0');
  return `${year}-${month}-${day}`;
}

function maybeAllowInsecureTls(apiBaseUrl) {
  if (apiBaseUrl.startsWith('https://localhost') || apiBaseUrl.startsWith('https://127.0.0.1')) {
    // Allow self-signed certs
    process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';
  }
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function get429DelayMs(attempt) {
  const exp = cfg.base429BackoffMs * 2 ** attempt;
  const jitter = Math.floor(Math.random() * 250);
  return Math.min(10000, exp + jitter);
}

async function retryFind(fn, attempts, sleepMs, label) {
  for (let i = 0; i < attempts; i += 1) {
    const value = await fn();
    if (value) {
      return value;
    }
    debug(`Retry ${i + 1}/${attempts} for ${label}...`);
    await sleep(sleepMs);
  }
  return null;
}

async function listAllPaginated(baseUrl, token, { size = 100 } = {}) {
  const out = [];
  for (let page = 1; page <= 20; page += 1) {
    const sep = baseUrl.includes('?') ? '&' : '?';
    const url = `${baseUrl}${sep}page=${page}&size=${size}`;
    const res = await authedJson(url, token, { allowError: true });
    if (!res.ok) {
      break;
    }
    const items = res.json?.items || [];
    out.push(...items);
    if (items.length < size) {
      break;
    }
  }
  debug(`listAllPaginated(${baseUrl}) -> ${out.length} items`);
  return out;
}

const green = '\x1b[32m';
const gray = '\x1b[90m';
const reset = '\x1b[0m';

function log(...args) {
  console.log(green + '[seed]', ...args, reset);
}

function debug(...args) {
  if (cfg.debug) {
    console.log(gray + '[seed:debug]', ...args, reset);
  }
}

function error(...args) {
  return new Error(`${args.join(' ')}`);
}

main().catch((err) => {
  console.error('[seed] Failed:', err.message);
  process.exit(1);
});
