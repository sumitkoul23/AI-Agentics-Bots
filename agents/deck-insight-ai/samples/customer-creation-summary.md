<!-- Deck Insight AI · persona: Head of Department · source: Customer_Creation_01_16.011.key -->
<!-- Example of the LLM-written briefing. The offline extractive mode produces a leaner version (see customer-creation-summary.offline.md). -->

## Executive Summary
This deck is a step-by-step Standard Operating Procedure for **creating a customer master record in SAP** using transaction **XD01**. It is aimed at the sales/commercial operations team who onboard new customers, and it matters because a customer cannot be invoiced or shipped to until this record is created correctly across all data views.

## What This Process Does
The procedure walks a user through SAP's centralised customer creation transaction (XD01), filling out the four data areas SAP requires — **General Data, Company Code Data, Control Data, and Sales Area Data** — and finishing with partner-function and extras settings. Done correctly, it produces a fully usable customer code that the business can sell to, ship to, and bill. The deck is clearly written for a regulated, GST-registered, pharmaceutical/drug-distribution context (it captures GST and Drug License numbers).

## Key Steps / Structure
1. Launch transaction **XD01**.
2. Enter the initial keys: Account Group **Z002**, Company Code **2100**, Sales Org. **Z210**.
3. **General Data** — customer name, address, and communication (mobile & email).
4. **Control Data** — GST number in *Tax Number 3*, Drug License number in *Tax Number 5*.
5. **Company Code Data** — Reconciliation Account **11000000** (fixed), Authorization **2100**.
6. **Sales Area Data → Sales** — Sales District, Sales Office, Sales Group, Customer Group, Currency.
7. **Sales Area Data → Shipping** — Delivery Priority, Shipping Conditions, Delivering Plant.
8. **Sales Area Data → Billing** — Incoterms = NA, Payment Terms, Acct. Assignment Group; GST code **0** if registered, **1** if not.
9. **Sales Area Data → Partner Functions** — assign the Sales Person.
10. **Extras → Customer Group 4** — **Z01** (Sold-to) / **Z02** (Ship-to).
11. Customer code is created.

## Critical Data & Configuration
- **Transaction:** XD01
- **Account Group:** Z002
- **Company Code:** 2100
- **Sales Org.:** Z210
- **Reconciliation Account:** 11000000 *(fixed — must not be changed)*
- **Authorization:** 2100
- **Tax Number 3:** GST Number  |  **Tax Number 5:** Drug License Number
- **GST registration code:** 0 = registered, 1 = not registered
- **Incoterms:** NA
- **Partner / Customer Group 4 codes:** Z01 = Sold-to, Z02 = Ship-to

## Risks, Compliance & Watch-outs
- **GST code inversion is the highest-risk field** — 0 vs 1 is easy to flip and directly drives tax treatment on every transaction with this customer.
- **GST number and Drug License number live in non-obvious fields** (Tax Number 3 and 5). Putting them in the wrong field is a silent compliance gap, especially for a regulated drug distributor.
- **Reconciliation Account 11000000 is fixed** — any deviation breaks financial posting integrity.
- **Sold-to vs Ship-to (Z01/Z02)** must match the real commercial relationship, or deliveries and invoices route to the wrong party.
- Missing Sales Area entries (district/office/group/currency) will block order creation downstream even though the customer "exists."

## Head of Department's Recommendations
- Convert this deck into a **one-page laminated checklist** with the fixed values pre-filled (Z002 / 2100 / Z210 / 11000000) so they are never re-typed from memory.
- Make **GST code (0/1), Tax Number 3, and Tax Number 5** mandatory double-check items in the onboarding review — ideally a second-person sign-off for regulated customers.
- Ask the SAP team to add **validation/derivation rules** for the fixed Reconciliation Account and for GST-code-vs-GST-number consistency.
- Run a **quarterly audit** of recently created customers against this SOP to catch field-placement and Sold-to/Ship-to errors early.
- Maintain a short **exceptions log** for any customer that deviates from the standard account group/org, so finance and compliance have a clear trail.
