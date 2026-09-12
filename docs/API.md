# URL Shortener API

Companion guide to [`openapi.yaml`](../openapi.yaml), which is the machine-readable source of truth.

Base URL (local): `http://localhost:8080`

---

## Contents

- [Quick start](#quick-start)
- [Authentication](#authentication)
- [How status works](#how-status-works) ← read this before building campaign UI
- [Endpoints](#endpoints)
- [Errors](#errors)
- [Generating TypeScript types](#generating-typescript-types)
- [Running the stack](#running-the-stack)

---

## Quick start

```bash
# 1. sign in via a browser (see Authentication), then copy the session cookie
export SESSION="<jwt from the session cookie>"

# 2. shorten a url
curl -X POST http://localhost:8080/api/urls \
  -H "Content-Type: application/json" -b "session=$SESSION" \
  -d '{"destination_url":"https://example.com/product/123","campaign_name":"instagram launch"}'

# 3. follow the short link
curl -i http://localhost:8080/OQ        # -> 302 Location: https://example.com/product/123

# 4. list your campaigns
curl -b "session=$SESSION" http://localhost:8080/api/urls
```

---

## Authentication

Google OAuth is the only sign-in method. There is no password login or registration endpoint.

**The flow**

1. Browser hits `GET /api/auth/google/login`. The server sets a 10-minute `oauth_state` cookie (CSRF protection) and 307s to Google.
2. The user consents. Google redirects to `GET /api/auth/callback?code=…&state=…`.
3. The server verifies `state` against the cookie, exchanges the code, fetches the Google profile, creates or refreshes the user, and sets a **`session`** cookie containing a JWT valid for **24 hours**.
4. The callback then **302s back to `FRONTEND_URL`** — it never returns a body. On failure it redirects to `FRONTEND_URL/login?error=<code>` (`login_cancelled`, `invalid_state`, `missing_code`, `login_failed`).
5. Every `/api/urls` request must carry that cookie.

**The JWT only ever appears in the `Set-Cookie` header.**

Two companion endpoints:

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/api/auth/me` | Current user, or 401. Lets a SPA bootstrap auth state — the cookie is httpOnly so JS can't read it. Treat 401 as "signed out", not an error. |
| `POST` | `/api/auth/logout` | Clears the cookie, 204. Unauthenticated on purpose, so a stale cookie can always be cleared. |

### CORS

The API allows exactly one origin, from `FRONTEND_URL`, with `Access-Control-Allow-Credentials: true`. A wildcard is not permitted alongside credentials, so any frontend on a different origin must be added there. Browser clients need `withCredentials: true` (axios) or `credentials: 'include'` (fetch), or the session cookie won't be sent.

### Getting a token for Postman / curl

The consent screen is interactive, so the flow cannot be completed from an API client. Do it once in a browser, then reuse the cookie:

1. Open `http://localhost:8080/api/auth/google/login` and sign in.
2. DevTools → **Application** → Cookies → `http://localhost:8080` → copy the `session` value.
   (Or DevTools → **Network** → the `callback` request → Response Headers → `Set-Cookie`.)
3. Send it as a header: `Cookie: session=<jwt>`, or add it to Postman's cookie jar for domain `localhost`.

Repeat daily, since the token expires after 24h. Replaying `/api/auth/callback` with a copied `code` will **not** work — the `oauth_state` cookie won't match and Google's code is single-use.

---

## How status works

**This is the detail most likely to trip up a frontend.**

`status` is **computed on every read** — it is not a field you set, except for pausing. The server derives it from the campaign's dates and whether the user disabled it:

| Returned `status` | When |
|---|---|
| `disabled` | The user paused it. **Always wins**, regardless of dates. |
| `scheduled` | `starts_at` is in the future. |
| `expired` | `expires_at` is in the past. |
| `active` | None of the above — it redirects right now. |

Only `active` campaigns redirect; the other three return 404 from `GET /{shortCode}`.

### What you may send

`PUT` accepts **`ACTIVE`** or **`DISABLED`** only (case-insensitive). These express intent: *enabled* or *paused*.

```jsonc
{"status": "DISABLED"}   // ok — pause it
{"status": "active"}     // ok — un-pause it
{"status": "SCHEDULED"}  // 400 — derived from starts_at, not settable
{"status": "EXPIRED"}    // 400 — derived from expires_at, not settable
```

To make a campaign *scheduled*, set a future `starts_at`. To *expire* it, set a past `expires_at`. To pause it regardless of dates, disable it.

### Enabling something expired

Allowed, and it simply reads back as `expired` — so you can enable and extend in one request:

```bash
# enable alone -> 200, but status is still "expired" and it won't redirect
curl -X PUT .../api/urls/OQ -d '{"status":"ACTIVE"}'

# enable AND extend -> status becomes "active" and it redirects
curl -X PUT .../api/urls/OQ -d '{"status":"ACTIVE","expires_at":"2027-01-01T00:00:00Z"}'
```

Validation applies to the **resulting** state, which is what makes the combined request work.

---

## Endpoints

All `/api/urls*` endpoints require the `session` cookie. `GET /{shortCode}` and the auth routes are public.

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/api/health` | Liveness check |
| `GET` | `/api/auth/google/login` | Start Google sign-in (browser only) |
| `GET` | `/api/auth/callback` | OAuth callback; sets the cookie, then 302s to the frontend |
| `GET` | `/api/auth/me` | Current user, or 401 |
| `POST` | `/api/auth/logout` | Clear the session cookie |
| `GET` | `/api/urls` | List your campaigns (`?search=&status=&limit=&offset=`) |
| `POST` | `/api/urls` | Shorten a URL (find-or-create the destination) |
| `GET` | `/api/urls/{shortCode}` | Fetch one campaign |
| `PUT` | `/api/urls/{shortCode}` | Update name / status / schedule |
| `POST` | `/api/destinations/{destinationId}/short-urls` | Add another campaign to a destination |
| `GET` | `/api/urls/{shortCode}/clicks` | Click analytics for a campaign |
| `GET` | `/{shortCode}` | Public redirect (302) + click tracking |

### `POST /api/urls` — two outcomes

This endpoint is find-or-create, so **check the status code**:

- **201** — the destination was new. Response has a single `short_url`.
- **200** — you already had this destination. Response has an array of its existing `short_urls`, and nothing was created. Any `campaign_name`/`starts_at`/`expires_at` you sent are **ignored**; to add another campaign use `POST /api/destinations/{destinationId}/short-urls`.

```jsonc
// 201
{ "destination_url": { "id": 5, "url": "https://example.com/product/123" },
  "short_url": { "id": 9, "code": "OQ", "url": "http://localhost:8080/OQ",
                 "campaign_name": "instagram launch", "status": "active" } }

// 200
{ "destination_url": { "id": 5, "url": "https://example.com/product/123" },
  "short_urls": [ { "id": 9, "code": "OQ", "url": "http://localhost:8080/OQ",
                    "campaign_name": "instagram launch", "status": "active" } ] }
```

### `GET /api/urls` — campaign list

Flat rows, newest first, with the destination embedded — maps straight onto a table.

```jsonc
{
  "campaigns": [
    { "id": 9, "code": "OQ", "short_url": "http://localhost:8080/OQ",
      "campaign_name": "instagram launch", "status": "active",
      "destination_url": "https://example.com/product/123",
      "starts_at": null, "expires_at": null,
      "created_at": "2026-09-06T02:33:48.043988+05:30" }
  ],
  "total": 11, "limit": 50, "offset": 0
}
```

`limit` defaults to 50 and caps at 100. Invalid `limit`, `offset` and `search` values fall back to defaults rather than erroring.

#### Filtering

Two optional filters narrow the list. **Both are applied in SQL, before pagination**, so `total` counts only matching campaigns and the page count stays correct.

| Param | Effect |
| --- | --- |
| `search` | Case-insensitive substring match against the campaign name, the short code **and** the destination url. A row matches if any of the three does. Whitespace-only is treated as absent; terms longer than 255 characters are truncated. |
| `status` | One of `active`, `scheduled`, `expired`, `disabled`. |

```bash
curl -b "session=$SESSION"   "http://localhost:8080/api/urls?search=uniqlo&status=active&limit=10"
```

`%` and `_` in a search term are matched **literally** — searching `50%` finds campaigns containing "50%", it does not match everything.

**`status` filters the derived status, not the stored column.** As described in [How status works](#how-status-works), only `active` and `disabled` are ever stored, so the server recomputes the effective status in the query itself. This is why `?status=expired` returns rows even though no row holds that value.

Unlike the other parameters, an unrecognised `status` is a **400** rather than a silent fallback — answering with the full unfiltered list would look like the filter was broken:

```jsonc
{ "error": "status must be one of active, scheduled, expired, disabled" }
```

### `PUT /api/urls/{shortCode}` — partial update

**Omitting a key leaves it unchanged. Sending `null` clears it.** That distinction is the whole point of the endpoint:

```jsonc
{"name": "renamed"}          // only the name changes
{"expires_at": null}         // campaign never expires  (and may flip expired -> active)
{"starts_at": null}          // starts immediately      (scheduled -> active)
{"name": null}               // clear the campaign name
{}                           // valid no-op, returns 200
```

A successful update **refreshes the redirect cache immediately** — disabling a campaign stops it redirecting on the very next request, not when a cache entry happens to expire.

Returns the full campaign under a `campaign` key, same shape as the list rows.

### `GET /api/urls/{shortCode}/clicks` — analytics

One endpoint, two shapes, chosen by whether `group_by` is present.

**Overview** (no `group_by`) — totals and the trend line:

```jsonc
{
  "range": { "start_ts": "2026-08-30T18:59:13Z", "end_ts": "2026-09-06T18:59:13Z" },
  "total_clicks": 35656,
  "series": [
    { "bucket": "2026-08-31T00:00:00Z", "clicks": 5027 },
    { "bucket": "2026-09-01T00:00:00Z", "clicks": 5090 }
  ]
}
```

**Breakdown** (`?group_by=country`) — one dimension at a time:

```jsonc
{
  "range": { "start_ts": "...", "end_ts": "..." },
  "group_by": "country",
  "total_clicks": 35655,
  "data": [
    { "value": "IN",      "clicks": 6044, "percentage": 16.95 },
    { "value": "unknown", "clicks":   20, "percentage":  0.06 }
  ]
}
```

`group_by` accepts **`country`, `city`, `region`, `device`, `browser`, `os`**. Anything else is a 400 listing the valid values. (`referrer` is intentionally absent — it stores raw URLs, so it needs host normalisation before grouping is useful.)

**Paging a breakdown**

Breakdowns are paged with `limit` (default 20, max 100) and `offset`, the same convention as `GET /api/urls`:

```bash
curl -b "session=$SESSION"   "http://localhost:8080/api/urls/OQ/clicks?group_by=city&limit=20&offset=20"   # page 2
```

The response echoes `limit`, `offset` and adds **`total_groups`** — the number of *distinct values* for that dimension:

```jsonc
{ "group_by": "city", "total_clicks": 504,
  "total_groups": 12, "limit": 20, "offset": 0,
  "data": [ { "value": "Mumbai", "clicks": 188, "percentage": 37.3 } ] }
```

Compute pages from `total_groups`, **not** `total_clicks` — the latter counts clicks, not rows, so 504 clicks across 12 cities would suggest 26 pages when there is only one:

```
offset     = limit * (page - 1)
page count = ceil(total_groups / limit)
```

These three fields appear only on a breakdown; an overview response has none of them.

**Time range**

| | |
|---|---|
| Default | last **7 days** |
| Maximum | **90 days** — a longer window is a 400, not a silent truncation |
| Params | `start_ts`, `end_ts` (RFC3339; `end_ts` defaults to now) |
| Buckets | hourly for ranges ≤ 2 days, daily beyond — keeps the series bounded |

```bash
curl -b "session=$SESSION" \
  "http://localhost:8080/api/urls/OQ/clicks?group_by=country&start_ts=2026-08-31T00:00:00Z&end_ts=2026-09-01T00:00:00Z"
```

**Three things that will otherwise surprise you**

1. **The `"unknown"` bucket is real data, not a bug.** Clicks with no value for the dimension are folded into it rather than dropped. Geo is unavailable for private/loopback IPs and OS is often unparseable, so on local data this bucket dominates.
2. **`data` rows on one page won't sum to 100%.** `percentage` is always a share of the full `total_clicks`, never of the current page — so page 2 of a breakdown might total 27%. Summing every page reaches 100%.
3. **The series only comes back on the overview.** Switching dimensions shouldn't recompute a trend the page already has, so breakdowns omit it.

Responses are cached in Redis for 60s per `(campaign, group_by, range)`. If Redis is down the endpoint still answers correctly from Postgres, just slower.

### `GET /{shortCode}` — the public redirect

- **302**, never 301 — deliberate, so repeat clicks keep reaching the server for analytics.
- Served from cache without a database read on a hit.
- Records the click asynchronously (user agent → browser/OS/device, IP → country/city/region). Analytics never delay or break the redirect.
- **404** for unknown codes *and* for `disabled`/`scheduled`/`expired` campaigns — indistinguishable by design.

---

## Errors

Every error is the same envelope:

```json
{ "error": "short url not found" }
```

| Status | Meaning | Common causes |
|---|---|---|
| `400` | Bad request | Malformed JSON; `status` not ACTIVE/DISABLED; `expires_at` not after `starts_at`; name over 255 chars; missing `destination_url` |
| `401` | Unauthorized | Missing, malformed or expired `session` cookie |
| `404` | Not found | Unknown short code, **or** it belongs to another user, **or** the campaign isn't currently redirecting |
| `500` | Server error | Database failure |

**On 404 and ownership:** "doesn't exist" and "belongs to someone else" return identical responses on purpose, so campaign ownership can't be probed.

---

## Generating TypeScript types

```bash
npx openapi-typescript openapi.yaml -o src/api-types.ts
```

Regenerate whenever `openapi.yaml` changes so the frontend can't silently drift from the backend.

---

## Running the stack

Three processes:

```bash
docker start url-shortener-redis   # cache + click queue
make run                           # api      (:8080)
make worker                        # click worker — writes clicks to postgres
```

The API still serves redirects if Redis or the worker is down; it falls back to Postgres and drops click events with a logged error. Migrations: `make migrate-up`.
