#!/usr/bin/env python3
"""
Simulate visitors clicking a short link, so the click pipeline and the
analytics dashboard can be exercised with realistic data.

Each request varies:
  * X-Forwarded-For  -> the server geolocates this against MaxMind GeoLite2
  * User-Agent       -> parsed into browser / os / device
  * Referer          -> stored as the click's referrer

Every IP and user agent below was verified against this project's own
GeoLite2 database and UA parser, so they produce real values rather than
"unknown" buckets.

Usage:
    python scripts/simulate_clicks.py --url http://localhost:8080/MjA --qps 5 --duration 100

The api, redis and the click worker all need to be running, or the events are
recorded nowhere:
    docker start url-shortener-redis
    make run
    make worker
"""

from __future__ import annotations

import argparse
import random
import statistics
import sys
import threading
import time
from collections import Counter
from concurrent.futures import ThreadPoolExecutor

import requests

# (ip, country, city) - resolved values confirmed against GeoLite2-City.mmdb.
LOCATIONS = [
    ("49.36.0.1", "IN", "Mumbai"),
    ("103.21.124.1", "IN", "Mumbai"),
    ("72.229.28.185", "US", "New York"),
    ("23.24.0.1", "US", "Philadelphia"),
    ("71.198.0.1", "US", "Livermore"),
    ("86.180.0.1", "GB", "Welshpool"),
    ("91.64.0.1", "DE", "Berlin"),
    ("133.11.0.1", "JP", "Tokyo"),
    ("14.200.0.1", "AU", "Melbourne"),
    ("189.6.0.1", "BR", "Brasilia"),
    ("24.48.0.1", "CA", "Montreal"),
    ("51.36.0.1", "SA", "Riyadh"),
]

# Weighted so the mix looks like a real campaign rather than a uniform spread:
# India-heavy for a UNIQLO India push, with a long tail elsewhere.
LOCATION_WEIGHTS = [26, 14, 9, 6, 5, 8, 6, 5, 5, 4, 4, 3]

# (label, user-agent) - each verified to parse into a real browser/os/device.
USER_AGENTS = [
    ("android-chrome", "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Mobile Safari/537.36"),
    ("samsung-chrome", "Mozilla/5.0 (Linux; Android 13; SM-S918B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Mobile Safari/537.36"),
    ("iphone-safari", "Mozilla/5.0 (iPhone; CPU iPhone OS 17_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.6 Mobile/15E148 Safari/604.1"),
    ("ipad-safari", "Mozilla/5.0 (iPad; CPU OS 17_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.6 Mobile/15E148 Safari/604.1"),
    ("win-chrome", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"),
    ("win-edge", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0"),
    ("win-firefox", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:133.0) Gecko/20100101 Firefox/133.0"),
    ("mac-safari", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.6 Safari/605.1.15"),
    ("linux-firefox", "Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0"),
    ("googlebot", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"),
]

# Mobile-dominant, as a social campaign would be, plus a little crawler traffic.
UA_WEIGHTS = [22, 13, 20, 4, 14, 5, 5, 8, 3, 2]

REFERRERS = [
    "https://www.instagram.com/",
    "https://t.co/",
    "https://www.facebook.com/",
    "https://www.google.com/",
    None,  # direct visit - no Referer header
]
REFERRER_WEIGHTS = [34, 22, 14, 12, 18]

# A slice of traffic arrives with no forwarded ip at all, which is what real
# unproxied requests look like. These land in the "unknown" geo bucket and are
# included on purpose so the dashboard shows honest coverage.
NO_GEO_RATE = 0.06


class Stats:
    def __init__(self) -> None:
        self.lock = threading.Lock()
        self.latencies_ms: list[float] = []
        self.statuses: Counter = Counter()
        self.countries: Counter = Counter()
        self.clients: Counter = Counter()

    def record(self, status: str, latency_ms: float, country: str, ua_label: str) -> None:
        with self.lock:
            self.statuses[status] += 1
            self.latencies_ms.append(latency_ms)
            self.countries[country] += 1
            self.clients[ua_label] += 1

    def sent(self) -> int:
        with self.lock:
            return sum(self.statuses.values())


def send_one(session: requests.Session, url: str, rng: random.Random, stats: Stats) -> None:
    ip, country, _city = rng.choices(LOCATIONS, weights=LOCATION_WEIGHTS, k=1)[0]
    ua_label, user_agent = rng.choices(USER_AGENTS, weights=UA_WEIGHTS, k=1)[0]
    referrer = rng.choices(REFERRERS, weights=REFERRER_WEIGHTS, k=1)[0]

    headers = {"User-Agent": user_agent}
    if rng.random() < NO_GEO_RATE:
        country = "unknown"
    else:
        headers["X-Forwarded-For"] = ip
    if referrer:
        headers["Referer"] = referrer

    started = time.perf_counter()
    try:
        # Never follow the redirect: we are measuring this service, not
        # spending the run fetching the destination site.
        response = session.get(url, headers=headers, allow_redirects=False, timeout=10)
        latency_ms = (time.perf_counter() - started) * 1000
        stats.record(str(response.status_code), latency_ms, country, ua_label)
    except requests.RequestException as exc:
        latency_ms = (time.perf_counter() - started) * 1000
        stats.record("error:" + type(exc).__name__, latency_ms, country, ua_label)


def percentile(values, pct: float) -> float:
    if not values:
        return 0.0
    ordered = sorted(values)
    index = min(len(ordered) - 1, int(round((pct / 100) * (len(ordered) - 1))))
    return ordered[index]


def main() -> int:
    parser = argparse.ArgumentParser(description="Simulate clicks on a short link.")
    parser.add_argument("--url", default="http://localhost:8080/MjA")
    parser.add_argument("--qps", type=float, default=5.0)
    parser.add_argument("--duration", type=float, default=100.0, help="seconds")
    parser.add_argument("--seed", type=int, default=None, help="set for a reproducible mix")
    args = parser.parse_args()

    rng = random.Random(args.seed)
    stats = Stats()
    total = int(round(args.qps * args.duration))
    interval = 1.0 / args.qps

    print("target   : " + args.url)
    print("rate     : {0:g} req/s for {1:g}s  (~{2} requests)".format(args.qps, args.duration, total))
    print("-" * 62)

    session = requests.Session()
    started_at = time.perf_counter()
    next_report = 10.0

    # A pool so one slow response cannot drag the send schedule behind.
    with ThreadPoolExecutor(max_workers=max(4, int(args.qps * 2))) as pool:
        for i in range(total):
            # Pace against the absolute start time so drift does not accumulate.
            target = started_at + i * interval
            sleep_for = target - time.perf_counter()
            if sleep_for > 0:
                time.sleep(sleep_for)

            pool.submit(send_one, session, args.url, rng, stats)

            elapsed = time.perf_counter() - started_at
            if elapsed >= next_report:
                print("  {0:5.1f}s  sent={1:4d}".format(elapsed, stats.sent()))
                next_report += 10.0

    elapsed = time.perf_counter() - started_at
    latencies = stats.latencies_ms

    print("-" * 62)
    print("done in {0:.1f}s   actual rate {1:.2f} req/s".format(elapsed, len(latencies) / elapsed))
    print("statuses      : {0}".format(dict(stats.statuses)))
    if latencies:
        print(
            "client latency: min {0:.1f}ms  median {1:.1f}ms  p95 {2:.1f}ms  max {3:.1f}ms".format(
                min(latencies),
                statistics.median(latencies),
                percentile(latencies, 95),
                max(latencies),
            )
        )
    print("countries sent: {0}".format(dict(stats.countries.most_common())))
    print("clients sent  : {0}".format(dict(stats.clients.most_common())))
    print("")
    print("Server-side redirect timings are in the api log")
    print('  ("redirection response time ms :- ...").')

    # Non-zero exit if anything other than a redirect came back, so this is
    # usable in a pipeline.
    return 0 if set(stats.statuses) <= {"302"} else 1


if __name__ == "__main__":
    sys.exit(main())
