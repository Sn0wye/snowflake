---
status: accepted
---

# Display-only FX: indicative unified total, settlement FX deferred

Each Account has a **Display Currency** preference, and the balance API returns an indicative **unified total** that sums all the Account's Wallets converted into that currency. This conversion is **display-only**: it never moves money, has no spread, and is independent of any Wallet's currency. **Settlement FX** — converting money inside a Transfer so a User can send USD and have the recipient receive EUR — remains out of scope.

## Considered Options

- **Rate source** — internal seeded table (no external dep but goes stale) vs. **external live FX API, cached with a TTL** (chosen, fresh rates) vs. defer the total entirely.
- **Fallback when rates are unavailable** — fail the balance call vs. omit the total whenever stale vs. **serve last-known rates flagged as stale, omit only on a cold cache** (chosen).

## Consequences

- Computing the unified total requires an outbound dependency on an FX provider, with a cache and TTL. This is acceptable only because the total is a convenience, not authoritative ledger data.
- The authoritative Wallet balances are **decoupled from FX uptime**: `GET /balance` always returns the Wallet list. When rates are past TTL the total is still computed from the last cached rates and flagged with an `as_of`/stale marker; when the cache has never warmed, the unified total is `null`. The balance read never fails because the FX provider is down.
- A future reader will see balances converted for display yet Transfers that refuse to cross currencies — that asymmetry is deliberate: display FX is read-only and tolerates staleness, settlement FX requires spread, rounding, and atomic money movement and is recorded as deferred (see [ADR-0003](0003-multi-wallet-accounts.md)).
