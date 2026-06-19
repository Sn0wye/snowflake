---
status: accepted
---

# Offset pagination for the transaction listing, not keyset

`GET /account/transactions` paginates by `page`/`limit` offset and returns a `total` count, rather than keyset/cursor pagination. Offset is the speculative-perf loser at scale — deep pages scan and discard skipped rows, and the `OR (sender OR receiver)` query needs a sort — but keyset would break the response contract (no `page` number, no jump-to-page-N, `total` becomes awkward) and require a `(created_at, id)` tiebreaker cursor since `created_at` isn't unique. The deep-page cost is bounded in practice (a hard 100-txn/account/day cap means years to accumulate a painful ledger, and reads are dominated by recent pages), so we keep the simpler offset contract and instead address the real cost with composite indexes `(sender_account_id, created_at)` and `(receiver_account_id, created_at)` that serve the directional (`credit`/`debit`) filters as index-ordered scans.

## Consequences

If deep paging or `COUNT(*)` ever becomes a measured bottleneck, revisiting this means a breaking contract change (cursor in, `page`/`total` out) — not a drop-in. That break is the price recorded here as deliberately deferred.
