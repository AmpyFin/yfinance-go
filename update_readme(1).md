
# update_readme.md — Add Web Scraping Fallback to `yfinance-go`

## 1) Problem & Motivation

Yahoo’s free endpoints sometimes **omit** features (e.g., **fundamentals/analysis/statistics**) or respond with **401/403**. Yet downstream Ampy systems still need these datasets for research and ops. This update adds a **first‑class, modular web‑scraping fallback** that activates **only when** the API path is unavailable, normalizes the data into **Ampy canonical contracts** (`ampy-proto`), and integrates with existing **CLI**, **bus**, **config**, and **observability**.

> Note: Scraping must respect operational guardrails. See **Legal/Operational Guardrails** below.

---

## 2) What This Adds (Feature Summary)

- **Scraper core**: hardened HTTP client with browser headers, gzip/deflate, rate limiting, retries/backoff, and session‑rotation reuse.
- **Embedded‑JSON extractors** for `key-statistics`, `financials` (income/balance/cashflow), `analysis`, `profile` by parsing the `quoteSummary` payloads embedded in `<script>` tags.
- **News HTML parser** for `/quote/{TICKER}/news/` (headlines, urls, sources, timestamps, related tickers, thumbnails).
- **Normalization to `ampy-proto`** (where available) with scaled decimals, currency codes, UTC time semantics, lineage/meta.
- **Fallback orchestration** that triggers scraping **only** upon API failure/denial for the specific endpoint, with per‑endpoint toggles.
- **CLI support**: `--fallback scrape` and `yfin scrape …` preview/export modes.
- **Config** via `ampy-config`: `scrape.enabled`, `scrape.qps`, `scrape.burst`, `scrape.user_agent`, `scrape.timeout`, `scrape.robots_policy`, `scrape.retry.*`.
- **Observability**: metrics/logs/traces for `scrape.fetch`/`scrape.parse`, success/error counters, latency histograms, page bytes, and decode failures.
- **Tests/Goldens**: fixtures for AAPL/MSFT/TSM to ensure stability and drift detection.

---

## 3) Success Criteria (Checklist)

- [ ] **Selective activation**: scraping happens only for endpoints that fail on API (e.g., fundamentals/analysis/statistics/profile/news) or when `--fallback scrape` is explicitly set.
- [ ] **Parity outputs**: scraped data normalizes to the **same `ampy-proto` contracts** used elsewhere; units, currency, timestamps validated.
- [ ] **Robust parsing**: extractor tests pass against fixtures; goldens guard mappings; graceful degrade on partial fields.
- [ ] **Throughput safety**: bounded concurrency + QPS/burst; exponential backoff with jitter; session rotation; no ban‑prone spikes.
- [ ] **Observability**: metrics/logs/traces present; dashboards show scraper health; alerts for elevated parse failures or 403/429 rates.
- [ ] **CLI UX**: `yfin scrape --ticker AAPL --endpoints financials,analysis --preview` prints redacted summaries; `--out json|parquet` works.
- [ ] **Bus (optional)**: when enabled, publish fundamentals snapshots and news raw records to configured topics with proper envelopes and ordering.
- [ ] **Config & Guardrails**: `ampy-config` schema extended and validated; robots/ToS policy surfaced; user-agent configurable.
- [ ] **No paid keys required**; this path remains **free‑tier** friendly.

---

## 4) Architecture (Docs‑only; no code yet)

```
cmd/yfin
  └── scrape/ (subcommand)           # will be added in later steps
internal/
  config/                            # extend with scrape.* later
  httpx/                             # reuse session rotation & backoff
  scrape/                            # new package to be added later
    client.go                        # HTTP client + headers + gzip + robots policy
    extractor_json.go                # <script data-url*="quoteSummary"> parsers
    extractor_news.go                # news HTML parser (defensive regex/selectors)
    normalize.go                     # shape→domain (scaled decimals, UTC, currency)
    types.go                         # internal DTOs for scraped shapes
  emit/                              # existing: map → ampy-proto
  bus/                               # existing optional publish
  obsv/                              # existing metrics/logs/traces
testdata/
  fixtures/yahoo/…                   # redacted pages for AAPL/MSFT/TSM
  golden/ampy/…                      # canonical outputs for regression
```

**Flow (fallback path, conceptual)**  
1) Library attempts API. On typed failure (401/403/404 or schema drift), controller invokes scraper for **that** endpoint.  
2) Scraper fetches HTML, extracts JSON (or article nodes), builds internal DTOs.  
3) Normalizer converts DTOs → `ampy-proto` messages (or JSON export for news/profile if proto not present).  
4) CLI either previews/exports or publishes (if `--publish` set).  
5) All steps emit logs/metrics/traces with `source="yfinance-go/scrape"`.

---

## 5) Endpoints Covered & Outputs (Docs‑only)

- **Key Statistics** → financial ratios/ownership → normalized into `ampy.fundamentals.v1` lines when representable; otherwise auxiliary JSON fields retained in export.
- **Financials** (IS/BS/CF) → quarterly/annual series → `ampy.fundamentals.v1.Snapshot` with `period_start/end`, currency, scaled values.
- **Analysis** → recommendation trends, earnings history → mapped into fundamentals/derived lines where applicable; the rest exported as JSON.
- **Profile** → `reference.v1.Company` if available; else JSON export + selected fundamentals lines (employees, sector).
- **News** → `news.v1.RawArticle` if schema exists; else JSON export and optional publish to `.../news/v1/raw` as opaque bytes.

> Currency: attach ISO‑4217, follow scaled‑decimal rules.  
> Time: all UTC; page relative times normalized with current clock at ingest.  
> Lineage: `meta.run_id`, `source="yfinance-go/scrape"`.

---

## 6) Legal & Operational Guardrails

- **Respect site policies**: A configurable `scrape.robots_policy` with default **enforce** will be introduced later; operators must consciously relax this if needed.  
- **Identify responsibly**: Configurable `User-Agent`.  
- **Rate limits**: conservative defaults (e.g., 0.5–1.0 QPS per host).  
- **No credential gating**: free‑tier only.  
- **Content stability**: extractor tests will pin layout expectations in later steps.

---

## 7) CLI Examples (Docs‑only)

```bash
# Use scraping only if API denies access (default behavior when enabled)
yfin scrape --ticker AAPL --endpoints financials,analysis --preview

# Force scraping (override API path) for debugging
yfin scrape --ticker MSFT --endpoints key-statistics,profile --preview --force

# Export results locally
yfin scrape --ticker TSM --endpoints news --out json --out-dir ./out

# Publish fundamentals from scraped financials (if bus enabled)
yfin scrape --ticker AAPL --endpoints financials --publish --env prod --topic-prefix ampy
```

---

## 8) Configuration Additions (`ampy-config`, docs‑only)

```yaml
scrape:
  enabled: true
  user_agent: "Mozilla/5.0 (Ampy yfinance-go scraper)"
  timeout_ms: 10000
  qps: 0.7
  burst: 1
  retry:
    attempts: 4
    base_ms: 300
    max_delay_ms: 4000
  robots_policy: "enforce"  # enforce|warn|ignore
  cache_ttl_ms: 60000       # optional HTML cache for in-run reuse
```

---

## 9) Testing Expectations (docs‑only)

- **Fixtures**: redacted HTML for AAPL/MSFT/TSM (news + quoteSummary pages).  
- **Unit**: extractors (JSON & news HTML) with drift cases; currency/time normalization.  
- **Integration**: fake HTTP server with gzip + 429/403 injection; controller fallback sequencing.  
- **Goldens**: canonical `ampy-proto` for financials; JSON for news/profile.  
- **CLI smoke**: `yfin scrape --ticker AAPL --endpoints financials --preview` prints expected summaries.

---

## If any problem arises…

Fix documentation **comprehensively** without adding out‑of‑scope implementation details. This step must remain strictly documentation‑only.
