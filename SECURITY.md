# Security Manual Test

## Pre-setup (account create)

### Services to run
- App stack: `docker compose up --build -d`
- App URL: `https://localhost:4443`

### Create admin account
1. Open `https://localhost:4443/register`
2. Create user with any valid values.
3. Verify the account.
4. In Mongo Express: `user-service` -> `users` -> edit your user:
- set `role` to `ADMIN`
5. Login at `https://localhost:4443/login`.

## XSS Injection

### Where to put value
- Page: `Admin -> Manage Genres`
- Action: `Create Genre`
- Field: `Genre Name`

### Values to try (copy list)
```txt
Demo<script>alert("XSS")</script>
Demo<img src=x onerror="alert('XSS')"/>
Demo<svg/onload=alert("XSS")>
Demo<a href="javascript:alert('XSS')">Click</a>
Demo&lt;script&gt;alert("XSS")&lt;/script&gt;
```

### Expected safe result
- No alert popup appears.
- No JavaScript executes.
- Input is either rejected or displayed as plain text.

## DoS attack

### Simple bash script (local demo)
Save as `dos-demo.sh` and run: `bash dos-demo.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

URL="https://localhost:4443/api/content/healthz"
WAVES=5
REQ_PER_WAVE=40

for w in $(seq 1 "$WAVES"); do
  echo "Wave $w/$WAVES"
  for i in $(seq 1 "$REQ_PER_WAVE"); do
    curl -k -s -o /dev/null -w "%{http_code}\n" "$URL" &
  done
  wait
  sleep 0.3
done

echo "Legitimate request after flood:"
curl -k -i "$URL"
```

### Expected safe result
- During burst, many requests should be rate-limited (`429`) if protection is active.
- Service should remain responsive (not crash).

## Other attacks

### SQL/NoSQL injection (stored)

Where to put value:
- Page: `Admin -> Manage Artists`
- Action: `Create Artist`
- Field: `Artist Name`

Values to try:
```txt
Demo{"$ne": null}
Demo{"$gt": ""}
Demo'; return true; var foo = '
Demo' OR '1'='1
Demo' UNION SELECT * FROM users--
Demo(a+)+$
```

Expected safe result:
- App does not crash.
- Payload is rejected or stored as literal text.
- No unauthorized data leak.

### Query injection (search endpoint)

Where to put value:
- Endpoint: `GET /api/content/search?q=<value>`
- Full URL base: `https://localhost:4443`

Values to try:
```txt
{"$ne":null}
{"$regex":".*"}
' OR '1'='1
'; DROP TABLE users;--
```

Quick curl examples:
```bash
curl -k "https://localhost:4443/api/content/search?q=%7B%22$ne%22:null%7D"
curl -k "https://localhost:4443/api/content/search?q=' OR '1'='1"
```

Expected safe result:
- Server returns controlled response (`200` normal JSON or `400` validation error).
- No full dump of all data.
- No 5xx crash loop.
