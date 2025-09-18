# Yahoo Finance → Ampy Proto Mapping Specification v1

**Scope:** Daily historical data only (1d intervals)

This document defines the unambiguous mapping from Yahoo Finance payloads to Ampy canonical messages (`ampy-proto`).

## Bars (1d historicals only)

| Yahoo field                      | `ampy.bars.v1` field                         | Notes |
|----------------------------------|----------------------------------------------|-------|
| `timestamp[i]`                   | `event_time` (end of bar); `start`,`end`     | Convert to **UTC**; for **1d** bars, `start` is 00:00:00Z of the day, `end` is next day 00:00:00Z; `event_time=end`. |
| `open[i]`,`high[i]`,`low[i]`    | `open`,`high`,`low`                          | Use **scaled decimals** `{scaled, scale}`; default equity scale `=4` (JPY examples may use `=2`). |
| `close[i]` (raw), `adjclose[i]` | `close`                                      | For **adjusted** bars use `adjclose`; for **raw** use `close`. |
| `volume[i]`                      | `volume`                                     | Integer. |
| `currency`                       | monetary fields `.currency_code`             | Always set (e.g., `USD`, `EUR`, `JPY`). No conversion here. |
| split/dividend metadata          | `adjusted`, `adjustment_policy_id`           | `split_dividend` for adjusted bars; `raw` for raw bars. |
| `meta.exchangeName`              | `security.mic` (derived)                     | Map to MIC when possible; otherwise leave empty and fill `exchange_name` in meta notes. |
| source                           | `meta.source="yfinance-go"`                  | Include `run_id`, `producer`, `schema_version`. |

## Quotes (snapshot)

| Yahoo field | `ampy.ticks.v1` field | Notes |
|-------------|-----------------------|-------|
| bid/ask     | `bid`,`ask`,`bid_size`,`ask_size` | Decimal with scale; sizes as integers. |
| symbol      | `security.symbol`     | MIC inferred as above. |
| ts          | `event_time`          | UTC. |
| venue       | `venue`               | Optional; set if present. |

## Fundamentals (quarterly)

| Yahoo field                 | `ampy.fundamentals.v1.Snapshot` | Notes |
|----------------------------|----------------------------------|-------|
| revenue, netIncome, eps    | `lines[].{key,value,currency_code}` | Use keys like `revenue`, `net_income`, `eps_basic`. |
| period start/end           | `period_start`, `period_end`     | UTC midnight boundaries. |
| as-of                      | `as_of`                          | UTC. |
| currency                   | `currency_code` at line-level    | Preserve source currency per line. |

## Notes

- **Daily only**: Yahoo Finance provides daily historical data only. Intraday intervals (1m/5m/1h) are out of scope.
- **Currency**: No conversion is performed. Source currency is preserved with explicit ISO-4217 codes.
- **Time semantics**: All timestamps are converted to UTC with proper daily boundaries.
- **Adjustments**: Both raw and adjusted data are supported with clear policy identification.
