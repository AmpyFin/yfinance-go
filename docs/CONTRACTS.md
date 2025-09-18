# CONTRACTS (Authoritative)

## Message FQDNs
- Bars: `ampy.bars.v1.Bar`, `ampy.bars.v1.BarBatch`
- Ticks: `ampy.ticks.v1.Tick` (QUOTE snapshots here)
- Fundamentals: `ampy.fundamentals.v1.FundamentalsSnapshot`

## Versioning
- Target: v1 across domains.
- Policy: additive only in v1; breaking → v2.

## Identity & Semantics (short)
- Timestamps UTC ISO-8601. Bars use [start inclusive, end exclusive].
- Decimals are scaled: `{scaled:<int>, scale:<int>}`.
- Monetary lines carry `currency_code` (ISO-4217).
- Bars include `adjusted` + `adjustment_policy_id`.
- Meta includes `run_id`, `source="yfinance-go"`, `producer`, `schema_version`.
