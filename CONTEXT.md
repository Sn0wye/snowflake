# Snowflake — Domain Glossary

## System Overview

Snowflake is a fintech platform where a **User**'s financial standing determines what they can do: their **Score** affects loan terms, their **Account** enables transfers, and their **FLAKE Keys** are how others address them for payments.

---

## Bounded Context: Identity (Helium)

### User
A registered person in the system. Owns a financial profile used by other contexts to make decisions. A User has exactly one Account and one Score.

### Financial Profile
The financial attributes of a User: `annual_income`, `debt`, and `assets_value` (all in USD). Owned by Helium. Read by Carbon to compute a Score and by Oxygen to evaluate loan eligibility.

---

## Bounded Context: Payments (Gold)

### Account
A ledger account associated with a User (1:1). Automatically provisioned when a User registers. Holds one or more **Wallets**, one per currency.

### Wallet
A per-currency holding within an Account. An Account has at most one Wallet per currency it holds (e.g. a USD Wallet and a EUR Wallet). Identified by the (Account, currency) pair. A Wallet holds a cached **Balance** in that currency's minor units.

### Balance
The cached balance of a single Wallet, in that Wallet's currency minor units. Authoritative value is the sum of that Wallet's ledger entries; the cached field exists for read performance and is validated by **Reconciliation**, which runs per Wallet.

### Account Status
The lifecycle state of an Account as a whole:
- `active` — default after provisioning; the Account may hold and transact through Wallets
- `closed` — set explicitly by the User; no further activity permitted

Suspension is per **Wallet**, not per Account (see **Wallet Status**).

### Wallet Status
The lifecycle state of a single Wallet:
- `active` — default; transfers in that currency are allowed
- `suspended` — set automatically when a reconciliation **Discrepancy** is detected for that Wallet; transfers in that currency are blocked until resolved, while the Account's other Wallets continue to operate

### Display Currency
A per-Account preference naming the currency in which the User wants their holdings summarised. Used only to compute an indicative unified total across all Wallets at an external **indicative FX rate**; it never moves money and is independent of any Wallet's own currency.

### Transfer
An atomic, **same-currency** payment between two Accounts. Identified by the recipient's **FLAKE Key** plus the currency being sent. Implemented as a simultaneous debit from the sender's Wallet of that currency and a credit to the receiver's Wallet of the same currency (auto-created if the receiver does not yet hold it). There is no pending or intermediate state — a Transfer either completes fully or does not occur. Cross-currency (FX-settled) transfers are not yet supported.

### FLAKE Key
A payment identifier owned by a User, used by others to address them as a Transfer recipient. Analogous to a PIX key. A User may register up to 5 FLAKE Keys. Supported types: `email`, `phone`, `cpf`, `cnpj`, `random` (UUID), `handle` (e.g., `@trzimajewski`).

### Reconciliation
The process of verifying that a **Wallet's** cached Balance equals the sum of that Wallet's ledger entries. Runs per Wallet (per (Account, currency) pair). A **Discrepancy** occurs when they diverge. Reconciliation is triggered periodically or on demand. On discrepancy, that Wallet is automatically suspended; the Account's other Wallets are unaffected.

### Discrepancy
The difference between a Wallet's cached Balance and its ledger-derived balance, in that Wallet's currency. Captured as an amount and a timestamp. Cleared when that Wallet's reconciliation passes again.

### Transaction
A recorded money movement against Accounts, denominated in a single currency. Today either a **Transfer** (same-currency, between two Accounts) or a **Deposit** (external funds into one Wallet). Atomic — a Transaction persists only once completed; there is no stored pending or failed Transaction. Distinct from the **Ledger Entry** it produces.

### Ledger Entry
The per-Wallet record of a Transaction's effect on one Wallet's Balance, in that Wallet's currency. A Transfer produces two Ledger Entries — a **debit** on the sender's Wallet and a **credit** on the receiver's Wallet, both in the same currency; a Deposit produces one credit. Immutable.

### Direction (credit / debit)
Whether a Transaction moved money into (**credit**) or out of (**debit**) a given Account. **Observer-relative**: the same Transfer is a debit to its sender and a credit to its receiver, so Direction is a property of the (Account, Transaction) pair — never of the Transaction alone. This is why a single credit/debit attribute cannot live on a Transaction.

---

## Bounded Context: Scoring (Carbon)

### Score
A creditworthiness measure for a User, ranging **0–900**. Calculated automatically on User registration and recalculated on demand. Composed of three sub-scores summed together.

### IncomeScore
Sub-score (0–300) derived from the User's annual income:
- ≥ $100k → 300
- $50k–$100k → 150–300 (linear)
- < $50k → 0–150

### AssetScore
Sub-score (0–300) derived from the User's assets value:
- ≥ $500k → 300
- $200k–$500k → 150–300 (linear)
- < $200k → 0–150

### DebtScore
Sub-score (0–300) derived from the User's debt-to-income ratio:
- ratio < 20% → 300
- ratio 20–40% → 150–300 (linear)
- ratio ≥ 40% → 0–150
- debt > income → 0

---

## Bounded Context: Loans (Oxygen)

### LoanApplication
A request by a User to borrow a specific amount for a specified **Term**. Evaluated automatically and immediately against the User's current **Score Tier** and **Income Cap**.

### LoanApplication Status
The lifecycle state of a LoanApplication:
- `PENDING` — initial state on submission
- `APPROVED` — requested amount is within the User's Income Cap for their Score Tier
- `REJECTED` — requested amount exceeds the User's Income Cap for their Score Tier

### Term
The duration of a loan, expressed in whole months (1–60).

### Score Tier
A named band of Score values that determines the **Base Rate** and **Income Cap** for a LoanApplication:

| Tier         | Score Range | Base Rate | Income Cap |
|--------------|-------------|-----------|------------|
| Poor         | 0–599       | 22%       | 15%        |
| Fair         | 600–674     | 16%       | 25%        |
| Good         | 675–724     | 12%       | 35%        |
| Very Good    | 725–774     | 8%        | 45%        |
| Excellent    | 775–849     | 5%        | 55%        |
| Outstanding  | 850–900     | 3%        | 65%        |

### Base Rate
The interest rate assigned to a Score Tier before term adjustment.

### Income Cap
The maximum loan amount a User may request, expressed as a percentage of their annual income, determined by their Score Tier.

### Term Multiplier
A factor applied to the Base Rate based on loan duration. Longer terms carry more lender risk:

| Term (months) | Multiplier |
|---------------|------------|
| 1–12          | 1.00×      |
| 13–24         | 1.15×      |
| 25–36         | 1.30×      |
| 37–48         | 1.50×      |
| 49–60         | 1.70×      |

### Final Rate
The effective interest rate for a LoanApplication: `Base Rate × Term Multiplier`.

---

## Cross-Context Events

### user.created
Published by Helium when a User registers. Consumed by:
- **Carbon** — triggers automatic Score calculation
- **Gold** — triggers automatic Account provisioning
