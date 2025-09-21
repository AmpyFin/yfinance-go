# yfinance-go — Yahoo Finance Client for Go (Open Source)

> ⚠️ **IMPORTANT DISCLAIMER** ⚠️
> 
> **This project is NOT affiliated with, endorsed by, or sponsored by Yahoo Finance or Yahoo Inc.**
> 
> This is an **independent, open-source Go client** that accesses publicly available financial data from Yahoo Finance's website. Yahoo Finance does not provide an official API for this data, and this client operates by scraping publicly accessible web pages.
> 
> **Use at your own risk.** Yahoo Finance may change their website structure at any time, which could break this client. We make no guarantees about data accuracy, availability, or compliance with Yahoo Finance's terms of service.
> 
> **Legal Notice:** Users are responsible for ensuring their use of this software complies with Yahoo Finance's terms of service and applicable laws in their jurisdiction.

---

> **Purpose:** A production‑grade **Go client** for Yahoo Finance data that:
> - Fetches **historicals** (bars), **quotes**, and **fundamentals** with correct **currency units**, **intervals**, and **corporate action adjustments**.
> - Produces **`ampy-proto`** payloads (e.g., `ampy.bars.v1.BarBatch`, `ampy.fundamentals.v1.Snapshot`) and, when configured, **publishes to `ampy-bus`** topics (e.g., `ampy/{env}/bars/v1/{mic}.{symbol}`).
> - Scales with **goroutines** (fan-out), **bounded concurrency**, and **rate‑limit backoff** to respect provider constraints.
>
> **Artifacts:** Library (Go) + **CLI** (`yfin pull --ticker AAPL --interval 1d`) for ops and CI.

---

## Installation

```bash
git clone https://github.com/yourusername/benzinga-golang.git
cd benzinga-golang
```

---

## 1) Mission & Success Criteria

**Mission**  
Provide a **reliable, consistent, and fast** Yahoo Finance client in Go that speaks **Ampy’s canonical contracts** (`ampy-proto`) and optionally **emits** to `ampy-bus`, so ingestion pipelines and research tools work identically across providers.

**Success looks like**
- Library returns **validated `ampy-proto` messages** with correct UTC times, currency semantics, and adjustment flags.
- CLI supports on-demand pulls and batch backfills; ops can **dry‑run**, **preview**, and **publish** with a single command.
- Concurrency and backoff keep **error rates** and **429/503** responses within policy; throughput is tunable and predictable.
- Golden samples round‑trip across **Go → Python/C++** consumers without shape drift.
- Observability shows latency/throughput, decode failures, and backoff behavior; alerts catch regressions.

---

## 2) Problems This Solves

- **Ad‑hoc yfinance usage** with drifting JSON shapes and ambiguous fields.
- **Float precision** and **currency mismatch** that break P&L and research.
- **Unbounded parallelism** causing bans/throttling.
- Missing **bus integration** and **ampy-proto** compliance.

---

## 3) Scope (What `yfinance-go` Covers)

- **Historical bars** (OHLCV) for equities/ETFs; **daily intervals only** (`1d`) - Yahoo Finance does not provide minute, hourly, weekly, or monthly historical data.
- **Quotes** (snapshot bid/ask/last, where available).
- **Fundamentals** (income/balance/cashflow lines, ratios) normalized into `ampy.fundamentals.v1`.
- **Corporate actions** awareness (splits/dividends) for adjusted vs raw bars.
- **Fan‑out** symbol fetches with **bounded goroutines**, **jittered exponential backoff**, and **rate‑limit windows**.
- **Optional publish** to `ampy-bus` using `ampy-bus` envelope conventions.
- **CLI** for operators, CI, and backfills.

**Non‑goals:** Options chains, intraday full‑depth order books, or broker execution. Those belong to other modules.

---

## 4) Global Conventions (Ingest Contract)

1) **Time**: All timestamps **UTC** ISO‑8601. Bars use `start` inclusive, `end` exclusive; `event_time` at bar close.  
2) **Precision**: Prices/amounts are **scaled decimals** (`scaled`, `scale`). Volumes are integers.  
3) **Currency**: Attach **ISO‑4217** code to monetary fields and fundamentals lines.  
4) **Identity**: Use `SecurityId` = `{ symbol, mic?, figi?, isin? }`. If MIC is unknown, prefer primary listing inference; document fallback rules.  
5) **Adjustments**: Bars declare `adjusted: true|false` and `adjustment_policy_id: "raw" | "split_only" | "split_dividend"`.  
6) **Lineage**: Every message has `meta.run_id`, `meta.source="yfinance-go"`, `meta.producer="<host|pod>"`, `schema_version`.  
7) **Batching**: Prefer `BarBatch` for efficiency. Maintain **in‑batch order** by `event_time` ascending.  
8) **Compatibility**: Additive evolution only; breaking changes require new major (`bars.v2`, `fundamentals.v2`).

---

## 5) Data Mappings (Yahoo → `ampy-proto`)

> Shapes below are **illustrative**. The library **must** handle provider quirks (missing fields, NaNs, market holidays).

### 5.1 Historical Bars → `ampy.bars.v1.BarBatch`

**Input (conceptual from Yahoo)**  
- `timestamp[]`, `open[]`, `high[]`, `low[]`, `close[]`, `adjclose[]`, `volume[]`, `currency`, `gmtoffset`, `validRanges`, `meta.exchangeName`

**Output (typical 1d adjusted bar)**
```json
{
  "security":{"symbol":"AAPL","mic":"XNAS"},
  "start":"2025-09-03T00:00:00Z",
  "end":"2025-09-04T00:00:00Z",
  "open":{"scaled":1931200,"scale":4},
  "high":{"scaled":1945000,"scale":4},
  "low":{"scaled":1926500,"scale":4},
  "close":{"scaled":1939800,"scale":4},
  "vwap":{"scaled":1939701,"scale":4},
  "volume":53498123,
  "trade_count":null,
  "adjusted":true,
  "adjustment_policy_id":"split_dividend",
  "event_time":"2025-09-04T00:00:00Z",
  "ingest_time":"2025-09-04T00:00:01Z",
  "as_of":"2025-09-04T00:00:01Z",
  "meta":{"run_id":"yfin_pull_20250904","source":"yfinance-go","producer":"ing-1","schema_version":"ampy.bars.v1:1.0.0"}
}
```

**Edge cases**
- Holidays/half‑days → bars absent or partial.
- Splits/dividends → adjusted vs raw track `adjclose` vs `close`.
- Currency differs from USD (`currency="JPY"`); ensure decimal scales per FX domain guidance.

### 5.2 Quotes → `ampy.ticks.v1` (snapshot as quote tick)

**Output example**
```json
{
  "security":{"symbol":"MSFT","mic":"XNAS"},
  "type":"QUOTE",
  "bid":{"scaled":4275000,"scale":4},
  "bid_size":200,
  "ask":{"scaled":4275300,"scale":4},
  "ask_size":300,
  "venue":"XNMS",
  "event_time":"2025-09-05T19:30:12Z",
  "ingest_time":"2025-09-05T19:30:12Z",
  "meta":{"run_id":"yfin_live_20250905","source":"yfinance-go","producer":"ing-2","schema_version":"ampy.ticks.v1:1.0.0"}
}
```

### 5.3 Fundamentals → `ampy.fundamentals.v1.Snapshot`

**Input (conceptual)**  
- Income/Balance/Cashflow items; currencies; period start/end; trailing vs quarterly.

**Output example (quarterly)**
```json
{
  "security":{"symbol":"AAPL","mic":"XNAS"},
  "lines":[
    {"key":"revenue","value":{"scaled":119870000000000,"scale":2},"currency_code":"USD","period_start":"2025-03-30T00:00:00Z","period_end":"2025-06-29T00:00:00Z"},
    {"key":"net_income","value":{"scaled":2386000000000,"scale":2},"currency_code":"USD","period_start":"2025-03-30T00:00:00Z","period_end":"2025-06-29T00:00:00Z"},
    {"key":"eps_basic","value":{"scaled":1525,"scale":2},"currency_code":"USD","period_start":"2025-03-30T00:00:00Z","period_end":"2025-06-29T00:00:00Z"}
  ],
  "source":"yfinance",
  "as_of":"2025-08-01T00:00:00Z",
  "meta":{"run_id":"yfin_fund_20250801","source":"yfinance-go","producer":"fund-1","schema_version":"ampy.fundamentals.v1:1.0.0"}
}
```

**Edge cases**: Restatements, negative EPS, mixed currency lines, missing items.

---

## 6) Concurrency, Rate Limits, and Backoff

- **Bounded goroutine pool**: max in‑flight requests configurable (e.g., 32, 64, 128).
- **Rate limits**: token bucket per host; configurable QPS and burst.
- **Backoff policy**: **exponential with jitter** on 429/5xx; max attempts and ceiling delay respected.
- **Fan‑out strategy**: group by exchange/MIC or by interval to maximize cache re‑use and reduce penalties.
- **Circuit breakers**: trip when consecutive error rate crosses threshold; auto half‑open probes.
- **Cold start warmup**: small initial concurrency to avoid immediate throttling.

**Telemetry (required)**
- Counters: `yfin.requests_total{outcome}`, `yfin.backoff_total{reason}`, `yfin.decode_fail_total{reason}`
- Histograms: `yfin.request_latency_ms`, `yfin.batch_size_bytes`
- Gauges: `yfin.inflight_requests`

---

## 7) Bus Integration (Optional Publish)

When enabled, the client publishes to `ampy-bus` with envelopes per `ampy-bus` spec.

**Topic examples**
- `ampy/prod/bars/v1/XNAS.AAPL`
- `ampy/prod/news/v1/raw` (not sourced here)
- `ampy/prod/fundamentals/v1/AAPL` (if a dedicated fundamentals topic is adopted; otherwise push to a data lake writer)

**Envelope headers (required)**
- `message_id` (UUIDv7), `schema_fqdn` (e.g., `ampy.bars.v1.BarBatch`), `schema_version`, `produced_at`, `producer`, `source="yfinance-go"`, `run_id`, `partition_key` (e.g., `XNAS.AAPL`), `trace_id`.

**Ordering & batching**
- Per `(symbol, mic)` ordering; maintain time‑ascending bars within batch.
- Chunk to keep **payload < 1 MiB** (pointer pattern if larger; rarely needed for daily bars).

---

## 8) CLI (Operator‑Facing Behavior)

> CLI is a thin wrapper over the library. It **prints** counts/summaries, can **preview** JSON payloads (redacted), and optionally **publishes**.

**Examples**
```bash
yfin pull --ticker AAPL --interval 1d --start 2024-01-01 --end 2024-12-31 --adjusted split_dividend --preview

yfin pull --universe-file nasdaq100.txt --interval 1d --start 2025-09-03 --end 2025-09-05 --publish --env prod --topic-prefix ampy/prod

yfin fundamentals --ticker AAPL --as-of 2025-08-01 --preview
```

**Flags (illustrative)**
- `--ticker`, `--universe-file`, `--market XNAS`
- `--interval 1d` (daily only - Yahoo Finance limitation)
- `--start`, `--end` (UTC)
- `--adjusted raw|split_only|split_dividend`
- `--publish`, `--env`, `--topic-prefix`
- `--concurrency`, `--qps`, `--retry-max`, `--backoff-max`
- `--preview` (print redacted JSON summaries), `--out parquet|json` (for local export)
- `--run-id` (explicit) else generated

**CLI Output (preview)**
- Counts per symbol, first/last timestamps, gaps detected, currency, and payload size summary.
- Error table with status codes and backoff applied.

---

## 9) Configuration (via `ampy-config`)

- **Transport**: base URLs, timeouts, compression, and cache TTLs.
- **Concurrency**: max inflight, QPS, burst per host.
- **Intervals**: daily (`1d`) only due to Yahoo Finance limitations; default adjustment policy.
- **Bus**: topic prefix, env, publish toggle.
- **Observability**: exporter endpoints, sampling, log level.
- **Secrets**: none required for public endpoints; if proxy/API key used, reference via secret URIs.

---

## 10) Validation & Testing

- **Golden samples**: For daily (`1d`) intervals across USD/EUR/JPY quotes; adjusted vs raw; fundamentals quarterly and trailing.
- **Round‑trip**: Serialize in Go, deserialize in Python using `ampy-proto`—values match exactly (including decimals).
- **Gap detection**: Known holiday windows produce expected absence; half‑days truncated correctly.
- **Backoff**: Inject 429/503; verify retry policy and counters.
- **Cardinality**: Metrics labels bounded; no `symbol` in labels (use logs/traces).
- **Performance**: Throughput at target concurrency without exceeding QPS/latency SLOs.
- **Regression**: Freeze “Yahoo schema → ampy-proto” mapping tests.

---

## 11) Observability & SLOs

- **Logs**: `bars.ingest`, `fundamentals.ingest`, `decode_fail`, with `run_id`, `trace_id`, `symbol`, `mic`.
- **Metrics**: Latency, throughput, failure counters; backoff totals.
- **Tracing**: Spans for `ingest.fetch`, `ingest.decode`, `bus.publish`.
- **SLO targets**:
  - p99 fetch latency (daily bars): **≤ 500 ms** per request (provider dependent)
  - p99 end‑to‑end ingest (fetch→publish) for daily bars: **≤ 1500 ms** under normal load
  - Error budget: 429/5xx rate **< 1%** sustained with backoff

---

## 12) Security & Compliance

- TLS for all network calls; proxy support.
- **No PII**; financial market data only.
- **Respect provider ToS**: rate limits, user‑agent, cache policy.
- Logs redact provider payloads by default (store hashes/byte sizes).

---

## 13) Failure Modes & Recovery

- **Provider outage**: throttle to minimum, surface alerts, skip publish.
- **Schema drift**: detect changed source fields → route to **DLQ topic** or error store; never silently coerce.
- **Clock skew**: trust provider timestamps; normalize to UTC; warn on future‑dated events.
- **Partial batches**: emit what’s valid; include `ingest_warnings` in meta if supported.

---

## 14) Acceptance Criteria (Definition of Done for `yfinance-go` v1)

1. Library returns **`ampy.bars.v1`** (adjusted/raw) and **`ampy.fundamentals.v1`** with full precision & correct time semantics.
2. Optional **publish** to `ampy-bus` using the canonical envelope; ordering by `(symbol, mic)` maintained.
3. CLI supports **single ticker** and **universe file** workflows with preview and publish modes.
4. Concurrency/backoff tested under stress; SLOs met; no rate‑limit violations in soak tests.
5. Golden samples locked; cross‑language round‑trips pass; mapping tests protect against upstream changes.
6. Observability: logs/metrics/traces emitted with correlation to bus messages; dashboards show ingest health.

---

## 15) End‑to‑End Narrative (Example)

1) Ops runs:
```bash
yfin pull --universe-file nasdaq100.txt --interval 1d --start 2024-01-01 --end 2024-12-31 --publish --env prod --run-id backfill_2024
```
2) Client fans out with 64 goroutines, respecting QPS. Bars are normalized (UTC, scaled decimals, currency) into `BarBatch`.  
3) Batches publish to `ampy/prod/bars/v1/{MIC}.{SYMBOL}` with UUIDv7 `message_id`, `run_id=backfill_2024`.  
4) Downstream **ampy-features** and data lake writers consume without adapters; Grafana shows ingest throughput and zero DLQ.  
5) A split detected for `TSLA` mid‑range: adjusted bars use `adjclose`; raw variant available by flag; both consistent with `corporate_actions.v1` semantics.

---

### Dependencies (as needed)
- **Ampy‑proto** (required): canonical schemas.
- **Ampy‑config** (recommended): typed config + secrets façade.
- **Ampy‑bus** (optional): publishing to NATS/Kafka.
- **Ampy‑observability** (recommended): logs/metrics/traces exporters & conventions.

> **Note on FX**: Real‑time currency conversion for fundamentals or non‑USD bars requires an FX source. For production we suggest choosing a provider (e.g., Open Exchange Rates, Currencylayer, ExchangeRate‑API). This may require an API key—see the step plan for where we’d integrate it.

---

## License
Apache‑2.0 (OS).

