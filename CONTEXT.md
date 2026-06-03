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
A ledger account associated with a User (1:1). Automatically provisioned when a User registers. Holds a cached **Balance** in USD cents.

### Balance
The cached USD-cent balance of an Account. Authoritative value is the sum of all ledger entries; the cached field exists for read performance and is validated by **Reconciliation**.

### Account Status
The lifecycle state of an Account:
- `active` — default after provisioning; transfers are allowed
- `suspended` — set automatically when a reconciliation discrepancy is detected; transfers are blocked until resolved
- `closed` — set explicitly by the User; no further activity permitted

### Transfer
An atomic payment between two Accounts. Identified by the recipient's **FLAKE Key**. Implemented as a simultaneous debit from the sender's Account and a credit to the receiver's Account. There is no pending or intermediate state — a Transfer either completes fully or does not occur.

### FLAKE Key
A payment identifier owned by a User, used by others to address them as a Transfer recipient. Analogous to a PIX key. A User may register up to 5 FLAKE Keys. Supported types: `email`, `phone`, `cpf`, `cnpj`, `random` (UUID), `handle` (e.g., `@trzimajewski`).

### Reconciliation
The process of verifying that an Account's cached Balance equals the sum of all its ledger entries. A **Discrepancy** occurs when they diverge. Reconciliation is triggered periodically or on demand. On discrepancy, the Account is automatically suspended.

### Discrepancy
The difference between an Account's cached Balance and its ledger-derived balance. Captured as an amount and a timestamp. Cleared when reconciliation passes again.

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
