---
status: accepted
---

# Multi-wallet accounts: one Account holds many per-currency Wallets

To support multiple currencies, an **Account** (still 1:1 with a User) now holds a set of **Wallets**, one per currency, each with its own cached **Balance** in that currency's minor units. The cached balance, reconciliation status, and suspension move off the Account and onto the Wallet, so a **Reconciliation** **Discrepancy** in one currency freezes only that Wallet while the Account's other Wallets keep transacting. Account-level status keeps only `active`/`closed`.

## Considered Options

- **One currency per Account** — rejected: a single User could then hold only one currency, which is not the feature.
- **Single base-currency Balance + FX-on-transaction** — rejected: collapses every currency into one number, loses per-currency ledgers, and forces settlement FX into every write path.
- **Multi-wallet (chosen)** — each currency is a first-class ledger with its own Balance, Ledger Entries, reconciliation, and suspension.

## Consequences

- A **Transfer** is **same-currency only**: it debits the sender's Wallet of currency X and credits the receiver's Wallet of the same currency, auto-creating the receiver's Wallet if absent so "a Transfer always completes" is preserved. Cross-currency (FX-settled) transfers are explicitly deferred; the schema (currency on every Transaction and Ledger Entry) is shaped so adding them later does not require a rewrite.
- Wallets are **provisioned lazily**: `user.created` creates an Account with zero Wallets; the first deposit/receive in a currency creates that Wallet. Reads must tolerate an Account holding no Wallets.
- Supported currencies are a **fixed allowlist carrying minor-unit exponent metadata**; balances stay `int64` in each currency's own minor units rather than assuming 2-decimal "cents".
- Per-transaction and daily **limits are removed entirely** in this slice and will be redesigned per-currency later.
- The balance read API becomes a list of Wallets rather than a single object — a breaking change to the previous single-Balance contract.
