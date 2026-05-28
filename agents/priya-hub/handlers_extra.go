package main

import "strings"

// ── Tax Strategist ────────────────────────────────────────────────────────────

func handleTax(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "capital gains"):
		return "Capital gains tax: short-term gains (assets held <1 year) are taxed as ordinary income. Long-term gains (>1 year) are taxed at 0%, 15%, or 20% depending on your income bracket. Consider tax-loss harvesting — selling losing positions to offset gains before year-end."
	case strings.Contains(lower, "tax loss harvest"):
		return "Tax-loss harvesting: sell investments at a loss to offset capital gains dollar-for-dollar. You can deduct up to $3,000 of net losses against ordinary income per year, carrying the rest forward. Beware the wash-sale rule — do not repurchase the same or substantially identical security within 30 days."
	case strings.Contains(lower, "s corp") || strings.Contains(lower, "llc tax"):
		return "S-Corp vs LLC tax: an S-Corp lets you split income into salary + distributions, reducing self-employment tax on the distribution portion. However it requires payroll, reasonable salary documentation, and extra compliance. An LLC taxed as a sole proprietorship is simpler but all profit is subject to SE tax. Run the numbers at ~$60K+ net profit to justify S-Corp election."
	case strings.Contains(lower, "deduction") || strings.Contains(lower, "write-off"):
		return "Common self-employed deductions: home office (exclusive use required), phone/internet (business %), vehicle mileage (67 cents/mile in 2024), health insurance premiums, retirement contributions (SEP-IRA up to 25% of net), software, education, and business travel. Keep receipts and a mileage log — the IRS will ask."
	case strings.Contains(lower, "irs") || strings.Contains(lower, "hmrc") || strings.Contains(lower, "tax return"):
		return "For US filers: quarterly estimated taxes are due April 15, June 15, September 15, January 15. Self-employed individuals owe SE tax (15.3% up to $168,600) plus income tax. UK HMRC Self Assessment deadline is January 31. Start early — organising receipts takes longer than the filing itself."
	default:
		return "For tax help, describe your situation: self-employed, employee, investor, or business owner? Country matters too (US, UK, EU rules differ significantly). I can cover deductions, entity structure, estimated payments, and capital gains planning."
	}
}

// ── Real Estate Advisor ───────────────────────────────────────────────────────

func handleRealEstate(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "cap rate") || strings.Contains(lower, "noi"):
		return "Cap rate = Net Operating Income / Property Value. A 5–7% cap rate is typical in urban markets; 8–10%+ signals higher risk or a secondary market. NOI = gross rents minus vacancy, operating expenses (taxes, insurance, maintenance, management), but before mortgage payments. Use cap rate to compare properties independently of financing."
	case strings.Contains(lower, "1031 exchange"):
		return "A 1031 exchange lets you defer capital gains tax by reinvesting proceeds from a sold investment property into a like-kind property. Key rules: identify replacement property within 45 days of sale, close within 180 days, use a qualified intermediary — proceeds cannot touch your hands. The deferred gain carries into the new property's basis."
	case strings.Contains(lower, "house hack"):
		return "House hacking: buy a 2–4 unit property, live in one unit, rent the others. The rental income offsets your mortgage — in strong rental markets it can cover it entirely. You qualify for owner-occupied financing (3.5–5% down FHA vs 20–25% for investment properties), dramatically improving returns. Run numbers at 75% occupancy to be conservative."
	case strings.Contains(lower, "rent vs buy"):
		return "Rent vs buy decision factors: price-to-rent ratio (purchase price / annual rent) — below 15 favours buying, above 20 favours renting. Also consider: how long you'll stay (buying wins at 5+ years), opportunity cost of the down payment, maintenance (budget 1–2% of value annually), and local market appreciation trends."
	case strings.Contains(lower, "reit") || strings.Contains(lower, "airbnb invest"):
		return "REITs offer real estate exposure without direct ownership — they must distribute 90% of taxable income as dividends. Publicly traded REITs are liquid; private REITs offer higher yields but illiquidity. For short-term rentals (Airbnb): gross yields can be 2–3x long-term rents, but factor in platform fees (3%), cleaning, furnishings, and higher management burden. Check local STR regulations first."
	default:
		return "For real estate advice, share the property type (residential, commercial, short-term rental), your market, investment budget, and goal (cash flow vs appreciation). I can model cap rates, financing scenarios, and tax implications."
	}
}

// ── Startup Coach ─────────────────────────────────────────────────────────────

func handleStartup(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "pitch deck"):
		return "A strong pitch deck (10–12 slides): Problem → Solution → Why Now → Market Size (TAM/SAM/SOM) → Product → Business Model → Traction → Team → Competition → Ask. Lead with traction — investors fund momentum. Every slide should answer one question clearly; remove anything that doesn't. What stage are you pitching?"
	case strings.Contains(lower, "cap table"):
		return "A clean cap table at seed: founders 70–80%, employees (option pool) 10–15%, seed investors 15–20%. Avoid over-diluting early — each round dilutes existing shareholders. Model your post-money ownership at Series A assuming a 20–25% dilution round. Use a SAFE or convertible note to defer valuation discussions at pre-seed."
	case strings.Contains(lower, "product market fit") || strings.Contains(lower, "mvp"):
		return "Product-market fit signals: organic word-of-mouth, users upset if product disappears (Sean Ellis >40% 'very disappointed' test), retention curves that flatten, NPS above 50. For MVP: build the smallest thing that delivers the core value proposition. Talk to 20+ potential customers before writing a single line of code."
	case strings.Contains(lower, "seed round") || strings.Contains(lower, "series a") || strings.Contains(lower, "raise capital") || strings.Contains(lower, "fundraise"):
		return "Seed round ($500K–$3M): typically SAFEs or convertible notes, pre-product or early traction. Series A ($3M–$15M): institutional VCs, requires $1M+ ARR with strong growth and unit economics. Warm intros convert 5–10x better than cold outreach. Build your investor pipeline like a sales pipeline — track stage, last contact, and next step."
	case strings.Contains(lower, "accelerator") || strings.Contains(lower, "term sheet"):
		return "Top accelerators (YC, Techstars, Founders Factory) offer capital, network, and credibility in exchange for 5–7% equity. Apply even if you don't need the money — the YC alumni network alone is worth it. On term sheets: watch valuation cap, pro-rata rights, information rights, and board composition. Have a startup lawyer review before signing."
	default:
		return "Tell me where you are in the journey: idea stage, building MVP, pre-revenue, or fundraising? I can help with pitch decks, cap table modelling, fundraising strategy, product-market fit, and go-to-market planning."
	}
}

// ── Sales Coach ───────────────────────────────────────────────────────────────

func handleSales(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "cold call"):
		return "Cold call framework: Permission opener ('Did I catch you at a bad time?') → One-line value prop tied to a pain you know they have → Single qualifying question → Ask for 15 minutes. Keep it under 60 seconds. The goal of a cold call is not to sell — it's to earn the discovery call. Personalise with one specific detail about their business before dialling."
	case strings.Contains(lower, "objection"):
		return "Objection handling (ACAF): Acknowledge ('I hear you, that's a fair concern') → Clarify ('Is it the price itself or the ROI uncertainty?') → Address with evidence → Flip ('Given that, does it make sense to test it on one team?'). Never argue. The most common objections — price, timing, need — are usually proxies for trust. Build trust first."
	case strings.Contains(lower, "discovery call"):
		return "Discovery call structure: Rapport (2 min) → Agenda setting → Situation questions (SPIN: Situation, Problem, Implication, Need-Payoff) → Confirm pain and impact → Preview next steps. Listen 70%, talk 30%. Your goal: understand their problem so precisely that your proposal writes itself. Record calls — you'll catch things you missed live."
	case strings.Contains(lower, "close the deal") || strings.Contains(lower, "crm pipeline"):
		return "Closing: ask for the next step at every stage, not just at the end. 'Based on what we covered, does it make sense to move forward?' The assumptive close works best — assume yes, lay out the path. In CRM, track: stage, deal size, next action, close date, and champion name. Deals without a clear champion almost never close."
	case strings.Contains(lower, "b2b sales") || strings.Contains(lower, "saas sales"):
		return "B2B/SaaS sales motion: Land with a small pilot (reduce risk perception) → Expand once they see value → standardise with an enterprise agreement. Average B2B sales cycle is 3–6 months; enterprise is 6–18 months. Multi-thread early — connect with multiple stakeholders. Economic buyer, technical buyer, and champion are all different people in deals above $50K."
	default:
		return "Describe your sales challenge: outbound prospecting, pipeline management, a specific deal, or building a sales process from scratch? I can walk through cold outreach scripts, discovery frameworks, objection handling, and closing techniques."
	}
}

// ── Marketing Strategist ──────────────────────────────────────────────────────

func handleMarketing(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "sales funnel") || strings.Contains(lower, "lead generation"):
		return "Funnel structure: Awareness (SEO, ads, content) → Consideration (email nurture, case studies, webinars) → Decision (free trial, demo, social proof) → Retention (onboarding, CS, upsell). Map your current drop-off by stage. Most businesses leak at Consideration — they generate leads but don't nurture them. What does your funnel look like?"
	case strings.Contains(lower, "cac ltv") || strings.Contains(lower, "customer acquisition"):
		return "CAC/LTV ratio: LTV should be at least 3× CAC for a healthy business. CAC = total sales + marketing spend / new customers acquired. LTV = average order value × purchase frequency × customer lifespan. To improve: reduce CAC (better targeting, referral programs) or increase LTV (upsells, retention, pricing). Which ratio is broken in your business?"
	case strings.Contains(lower, "seo strategy"):
		return "SEO foundation: technical (site speed, Core Web Vitals, crawlability) → on-page (keyword research, content clusters, E-E-A-T) → off-page (backlinks from relevant sites). Prioritise long-tail keywords with commercial intent first — easier to rank and closer to conversion. Content clusters (pillar page + supporting articles) outperform isolated posts. What's your domain authority and existing content?"
	case strings.Contains(lower, "email campaign"):
		return "Email performance benchmarks: open rate 20–30% (B2B), click rate 2–5%, unsubscribe <0.5%. Improve opens: subject line A/B tests, send-time optimisation, list hygiene. Improve clicks: single CTA, personalisation, mobile-first design. Automated sequences (welcome, nurture, re-engagement) outperform one-time blasts 3–4× in revenue per email."
	case strings.Contains(lower, "a/b test") || strings.Contains(lower, "growth hack"):
		return "A/B testing: only change one variable at a time. Minimum sample size per variant: 1,000 conversions for statistical significance (use a calculator). Prioritise tests by ICE score: Impact × Confidence × Ease. Don't stop tests early — let them run to significance. Landing page headline, CTA copy, and pricing page are highest-impact test areas."
	default:
		return "Describe your marketing challenge: customer acquisition, funnel leaks, content strategy, paid ads, or brand building? Share your product, target customer, and current channels. I'll give you a prioritised growth framework."
	}
}

// ── Legal Advisor ─────────────────────────────────────────────────────────────

func handleLegal(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "nda"):
		return "NDA essentials: define 'Confidential Information' precisely (don't use blanket language), specify permitted disclosures, set a clear term (1–3 years typical), include return/destruction of materials, and state governing law. One-way NDAs favour the disclosing party; mutual NDAs are standard between parties sharing information both ways. Always have a lawyer review before signing anything binding. This is general guidance, not legal advice."
	case strings.Contains(lower, "trademark") || strings.Contains(lower, "intellectual property"):
		return "Trademark: register your brand name and logo with the USPTO (US) or EUIPO (EU). Use TM before registration, ® after. Classes matter — registration covers specific goods/services categories. The process takes 8–12 months; expect ~$250–$400 per class. Common-law rights exist from first use but give weaker protection. Search the TESS database before adopting a new mark. This is general guidance, not legal advice."
	case strings.Contains(lower, "contract"):
		return "Key contract clauses to review: payment terms, IP ownership (especially for freelancers — does work transfer to client or remain yours?), limitation of liability, termination provisions, dispute resolution (arbitration vs litigation, governing jurisdiction). Red flags: unlimited liability, no termination for convenience, IP assignment with no carve-outs. Always negotiate — most contracts are templates. This is general guidance, not legal advice."
	case strings.Contains(lower, "gdpr") || strings.Contains(lower, "privacy policy"):
		return "GDPR compliance basics: identify all personal data you collect, establish a lawful basis for processing (consent, legitimate interest, contract), maintain a data processing register, appoint a DPO if required, implement breach notification within 72 hours, and honour data subject rights (access, erasure, portability). Fines reach €20M or 4% of global revenue. Get a GDPR-specialised lawyer if you serve EU users. This is general guidance, not legal advice."
	case strings.Contains(lower, "incorporation") || strings.Contains(lower, "patent") || strings.Contains(lower, "copyright"):
		return "Incorporation: Delaware C-Corp is standard for VC-backed startups (flexible equity, investor familiarity). LLC is simpler for small businesses and pass-through taxation. Copyright protects original works automatically on creation; registration (US Copyright Office, ~$65) is required to sue for statutory damages. Patents require filing before public disclosure — provisionals buy you 12 months. This is general guidance, not legal advice."
	default:
		return "Describe your legal question: contract review, entity formation, IP protection, employment law, or regulatory compliance? Note: I provide general legal information only — for binding decisions, consult a qualified attorney in your jurisdiction."
	}
}

// ── HR & People ───────────────────────────────────────────────────────────────

func handleHR(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "job description") || strings.Contains(lower, "hiring plan"):
		return "Strong job descriptions: lead with impact ('You will own X and deliver Y') not just duties, be explicit about seniority expectations (avoid 'rockstar'), list 5 must-haves not 15 nice-to-haves, include comp range (reduces unqualified applications by 30%), and state the interview process upfront. A hiring plan should map headcount to business milestones, not a fixed calendar."
	case strings.Contains(lower, "interview question") || strings.Contains(lower, "onboard employee"):
		return "Structured interviews outperform unstructured ones 2× in predictive validity. Use the same questions for all candidates and score with a rubric before discussing. Best questions: 'Tell me about a time you [specific challenge relevant to role]' — force specifics, not hypotheticals. For onboarding: 30/60/90 day plan with clear milestones, assigned buddy, early win scoped within first 2 weeks."
	case strings.Contains(lower, "performance review") || strings.Contains(lower, "compensation"):
		return "Performance reviews: connect ratings to concrete outcomes, not effort or personality. Use a calibration session across managers before sharing with employees. Compensation bands: research market rates (Levels.fyi, Radford, Glassdoor), set bands by role and level, review annually. Pay transparency within bands reduces inequity and retention risk. Never give COLA increases as a substitute for performance raises."
	case strings.Contains(lower, "equity plan"):
		return "Employee equity: standard 4-year vest with 1-year cliff is the baseline. Options (ISOs for US employees, EMIs in the UK) are most common at startups. RSUs make more sense post-Series B when there is a clear liquidity path. Communicate total compensation including equity value clearly — most employees undervalue their options because they don't understand them."
	case strings.Contains(lower, "fire employee") || strings.Contains(lower, "remote team") || strings.Contains(lower, "team culture"):
		return "Terminations: document performance issues before the conversation (PIPs for performance, separate process for conduct). Have HR or legal counsel present. Give clear, factual reasons — vague 'not a fit' language increases litigation risk. For remote teams: over-communicate async, set explicit working hours expectations, create deliberate social rituals, and measure outputs not hours."
	default:
		return "Describe your people challenge: hiring, compensation, performance management, equity design, culture, or a specific employee situation. I can provide frameworks and templates for each stage of the employee lifecycle."
	}
}

// ── E-Commerce Expert ─────────────────────────────────────────────────────────

func handleEcommerce(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "shopify"):
		return "Shopify setup priorities: choose a fast theme (Dawn is solid), configure Shopify Payments before third-party gateways to avoid transaction fees, install only essential apps (every app slows your store), set up abandoned cart emails (recovers 5–15% of lost revenue), and enable Shop Pay — it increases conversion by ~50% for returning customers. What specifically are you building or optimising?"
	case strings.Contains(lower, "amazon fba"):
		return "Amazon FBA key metrics: ACoS (Advertising Cost of Sale) should be below your net margin. Target a BSR (Best Seller Rank) in the top 1% of your subcategory to generate organic traffic. Product research: look for >300 units/month sold, <50 reviews on page-1 listings, and a selling price above $25 to leave room for FBA fees (~$3–$5 per unit) and PPC. Private label margins after all fees: target 25–35%."
	case strings.Contains(lower, "dropshipping"):
		return "Dropshipping reality check: margins are 15–30% vs 40–70% for private label. Competing on price is a race to the bottom. Differentiate through branding, superior product pages, and customer service. Best niches are passion-based (pets, fitness, hobbies) with repeat purchase potential. Supplier reliability is everything — use AliExpress for testing, then move to domestic suppliers for faster shipping once validated."
	case strings.Contains(lower, "product listing") || strings.Contains(lower, "etsy"):
		return "Product listing optimisation: main image (white background, product filling 85% of frame) drives CTR. Title: lead with the primary keyword, follow with attributes (size, material, colour). Bullet points: lead each with a benefit, back it with a feature. A+ Content (Amazon) or video (Shopify) increases conversion 5–10%. On Etsy: keyword research with eRank, use all 13 tags, photos with lifestyle context outperform white backgrounds."
	case strings.Contains(lower, "fulfillment") || strings.Contains(lower, "inventory"):
		return "Inventory management: calculate reorder point = (average daily units × lead time) + safety stock. Safety stock = z-score × standard deviation of demand × lead time. Too much inventory ties up cash; too little creates stockouts and lost BSR. For 3PL vs FBA: FBA wins on Prime badge and discovery; 3PL wins on cost for large, heavy, or multi-channel products."
	default:
		return "Tell me your e-commerce situation: what platform, what product category, and what's the specific challenge — traffic, conversion, supplier, fulfillment, or profitability? I can model unit economics, ad strategies, and scaling plans."
	}
}

// ── DevOps Engineer ───────────────────────────────────────────────────────────

func handleDevops(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "terraform"):
		return "Terraform best practices: store state in remote backend (S3 + DynamoDB lock, or Terraform Cloud), use workspaces or separate state files per environment, pin provider versions, and use modules for reusable infrastructure. Follow the plan → review → apply workflow in CI — never apply locally in production. Use `terraform fmt` and `tflint` in pre-commit hooks. What infrastructure are you provisioning?"
	case strings.Contains(lower, "helm chart") || strings.Contains(lower, "container orchestration") || strings.Contains(lower, "kubernetes"):
		return "Helm chart structure: Chart.yaml → values.yaml → templates/. Use `helm lint` and `helm template` to validate before deploying. Set resource requests and limits on every container — unset limits cause noisy-neighbour problems. Use Horizontal Pod Autoscaler with CPU/memory thresholds. Separate values files per environment (values-prod.yaml). Consider ArgoCD or Flux for GitOps-style deployments."
	case strings.Contains(lower, "github actions pipeline") || strings.Contains(lower, "jenkins") || strings.Contains(lower, "ci/cd"):
		return "CI/CD pipeline stages: lint → unit tests → build → integration tests → security scan (Trivy, Snyk) → push image → deploy to staging → smoke test → promote to production. Use environment protection rules and required reviewers for production deploys. Cache dependencies (npm, Go modules, pip) to cut build times 50–80%. Fail fast — put fastest checks first."
	case strings.Contains(lower, "prometheus") || strings.Contains(lower, "grafana") || strings.Contains(lower, "site reliability") || strings.Contains(lower, "sre"):
		return "SRE fundamentals: define SLOs (e.g., 99.9% availability, p99 latency <200ms), then SLIs to measure them, then error budgets to govern release velocity. Prometheus scrapes metrics; Grafana visualises. Alert on symptoms not causes — alert when SLO burn rate is high, not on individual server CPU. On-call runbooks should be actionable in under 5 minutes by anyone on the team."
	case strings.Contains(lower, "ansible") || strings.Contains(lower, "cloud architecture"):
		return "Cloud architecture principles: design for failure (multi-AZ by default), use managed services where possible to reduce operational burden, separate stateless and stateful layers, and implement least-privilege IAM. Ansible excels at configuration management and ad-hoc tasks; use roles and inventory groups for maintainability. Avoid snowflake servers — everything should be reproducible from code."
	default:
		return "Describe your DevOps challenge: infrastructure provisioning, CI/CD pipeline, container orchestration, observability, or incident response? Share your stack (cloud provider, orchestrator, languages) and I'll give specific, actionable guidance."
	}
}

// ── Data Scientist ────────────────────────────────────────────────────────────

func handleData(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "machine learning model") || strings.Contains(lower, "pandas") || strings.Contains(lower, "numpy"):
		return "ML workflow: data collection → EDA (pandas-profiling, visualisations) → feature engineering → model selection → hyperparameter tuning → evaluation → deployment → monitoring. Start with the simplest model that could work (linear regression, logistic regression). Only add complexity when baseline performance is insufficient. Always establish a random-guess baseline first. What's your target variable and dataset size?"
	case strings.Contains(lower, "data pipeline") || strings.Contains(lower, "etl pipeline"):
		return "ETL pipeline design: Extract (API, DB, files) → Transform (validation, cleaning, joins, aggregations) → Load (data warehouse, lake). Use idempotent transformations so rerunning doesn't corrupt data. Tools: dbt for SQL transformations (version-controlled, testable), Airflow or Prefect for orchestration, Great Expectations for data quality assertions. Incremental loads beat full refreshes for tables over 1M rows."
	case strings.Contains(lower, "bigquery") || strings.Contains(lower, "snowflake") || strings.Contains(lower, "data warehouse"):
		return "Data warehouse modelling: use a dimensional model (Kimball) with fact tables (events, transactions) and dimension tables (users, products, dates). Partition fact tables by date to reduce query cost. In BigQuery: use clustered columns on high-cardinality filter fields, avoid SELECT *, and cache results with materialised views. In Snowflake: size your warehouse to the query pattern — use multi-cluster for high concurrency."
	case strings.Contains(lower, "looker") || strings.Contains(lower, "tableau") || strings.Contains(lower, "analytics dashboard"):
		return "Dashboard design principles: one metric per chart, clear titles that state the insight not just the variable, consistent colour coding (red = bad, green = good unless colour-blind considerations apply), and a single 'north star' metric at the top. In Looker: define metrics in LookML so they stay consistent everywhere. Avoid dashboards with more than 8–10 charts — they become wallpaper. Who is the audience and what decision does this dashboard drive?"
	case strings.Contains(lower, "jupyter"):
		return "Jupyter notebook best practices: use nbstripout to strip outputs before committing to git, parameterise notebooks with Papermill for scheduled runs, use nbconvert to generate shareable reports. For production: convert notebooks to scripts or use Prefect/Airflow. Structure notebooks: imports → config → data load → EDA → modelling → evaluation → conclusions. Use markdown cells to explain findings, not just code."
	default:
		return "Describe your data challenge: pipeline design, model building, warehouse modelling, dashboard creation, or statistical analysis? Share the data source, scale (rows, update frequency), and the business question you're trying to answer."
	}
}

// ── Security Expert ───────────────────────────────────────────────────────────

func handleSecurity(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "penetration test") || strings.Contains(lower, "pentest"):
		return "Penetration testing phases: Reconnaissance (passive: OSINT, shodan; active: port scanning) → Scanning & Enumeration → Exploitation → Post-Exploitation → Reporting. Always get written authorisation before testing. Scope must be explicit — IP ranges, domains, out-of-bounds systems. Report format: executive summary, risk-rated findings (CVSS scores), reproduction steps, and remediation recommendations. What's your target environment?"
	case strings.Contains(lower, "sql injection attack") || strings.Contains(lower, "xss attack") || strings.Contains(lower, "owasp"):
		return "OWASP Top 10 mitigations: SQLi → parameterised queries (never string concatenation). XSS → output encoding, Content-Security-Policy header. IDOR → server-side authorisation checks on every request. SSRF → allowlist outbound destinations. Broken auth → MFA, short-lived tokens, secure session storage. Run OWASP ZAP or Burp Suite against your app before every major release. Which vulnerability are you investigating?"
	case strings.Contains(lower, "threat model"):
		return "Threat modelling (STRIDE): Spoofing, Tampering, Repudiation, Information Disclosure, Denial of Service, Elevation of Privilege. Process: define assets → draw data-flow diagrams → identify threats per DFD element → rate by likelihood × impact → prioritise mitigations. Threat modelling is most valuable done early in design, before code is written. Tools: OWASP Threat Dragon, Microsoft Threat Modeling Tool."
	case strings.Contains(lower, "incident response"):
		return "Incident response lifecycle: Preparation → Detection & Analysis → Containment → Eradication → Recovery → Post-Incident Review. First 30 minutes: isolate affected systems, preserve logs (do not wipe), notify legal/comms if data breach suspected. GDPR requires breach notification within 72 hours of discovery. Document everything with timestamps. Run tabletop exercises quarterly so the process is muscle memory."
	case strings.Contains(lower, "security audit") || strings.Contains(lower, "red team") || strings.Contains(lower, "vulnerability scan"):
		return "Vulnerability scanning tools: Nessus/Tenable, OpenVAS for network/host scanning; Trivy, Snyk, Grype for container and dependency scanning; Semgrep, Bandit for static code analysis. Run scans in CI — shift security left. Prioritise CVSS 7+ findings. Red team vs pentest: pentests are time-boxed and scoped; red team engagements simulate full APT campaigns with no predefined scope. What's your security maturity level?"
	default:
		return "Describe your security question: a specific vulnerability, security architecture review, compliance requirement (SOC 2, ISO 27001, PCI-DSS), or incident you're responding to. I can walk through technical mitigations, processes, and tooling."
	}
}

// ── Web3 Developer ────────────────────────────────────────────────────────────

func handleWeb3(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "smart contract") || strings.Contains(lower, "solidity"):
		return "Solidity security checklist: reentrancy (use checks-effects-interactions pattern, or ReentrancyGuard), integer overflow (Solidity 0.8+ reverts by default), access control (OpenZeppelin Ownable/AccessControl), front-running (commit-reveal or Chainlink VRF for randomness), and oracle manipulation (use TWAPs not spot prices). Audit with Slither, Mythril before deployment. What contract are you building?"
	case strings.Contains(lower, "defi protocol"):
		return "DeFi protocol design: define the core primitive (AMM, lending, yield aggregator), model the token economics and incentive alignment carefully, implement price oracle security (Chainlink TWAP), add circuit breakers for extreme price movements, and plan the upgrade path (transparent proxy vs UUPS). The biggest DeFi exploits all came from oracle manipulation or reentrancy — audit both exhaustively."
	case strings.Contains(lower, "nft project") || strings.Contains(lower, "dao governance") || strings.Contains(lower, "tokenomics"):
		return "NFT launch: use ERC-721A (batch minting saves 70%+ gas vs ERC-721), implement a merkle-tree whitelist, store metadata on IPFS with a fallback, and plan reveal mechanics carefully (pre-reveal hash commitment to prevent sniping). For DAOs: Governor Bravo + Timelock is battle-tested. Tokenomics: avoid linear vesting cliffs — stepped vesting with lockup aligns long-term incentives better."
	case strings.Contains(lower, "layer 2") || strings.Contains(lower, "polygon network") || strings.Contains(lower, "arbitrum"):
		return "L2 selection: Arbitrum One (Nitro) and Optimism use optimistic rollups — 7-day withdrawal period, lower costs, EVM-equivalent. Polygon PoS is a sidechain (not a true L2) — faster finality but different security assumptions. ZK rollups (zkSync, Starknet, Polygon zkEVM) offer faster finality and stronger security but have higher proving costs. Choose based on your latency needs, security requirements, and which has your target user base."
	case strings.Contains(lower, "zk proof") || strings.Contains(lower, "hardhat") || strings.Contains(lower, "foundry framework"):
		return "Foundry vs Hardhat: Foundry (Rust-based, forge/cast) is faster for testing and has built-in fuzzing — prefer it for new projects. Hardhat has a larger plugin ecosystem and is more familiar to JS developers. ZK proofs (zk-SNARKs via Circom/snarkjs, or zk-STARKs via StarkWare) prove computation without revealing inputs. Use ZKPs for privacy-preserving features (proof of age, proof of balance) not general compute — they're expensive to generate."
	default:
		return "Describe your Web3 project: which chain, what the contract does, and where you are in the development lifecycle (design, build, testing, audit, deployment). I can help with Solidity, protocol economics, security, and L2 architecture."
	}
}

// ── Writing Coach ─────────────────────────────────────────────────────────────

func handleWriting(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "copywriting") || strings.Contains(lower, "sales copy") || strings.Contains(lower, "ad copy"):
		return "High-converting copy formula: PAS (Problem → Agitate → Solve) or AIDA (Attention → Interest → Desire → Action). Lead with the reader's pain, not your product's features. One CTA per piece — multiple options kill conversion. Social proof near the CTA (testimonials, numbers, logos). Power words: 'you', 'proven', 'instant', 'guarantee', 'because'. What product or service are you writing for, and who is the reader?"
	case strings.Contains(lower, "blog post writing") || strings.Contains(lower, "long-form"):
		return "Blog post structure: Hook (why should I read this?), Promise (what will I learn?), Body (deliver the promise in scannable sections with H2/H3), Proof (data, examples, stories), and CTA. Aim for 1,500–2,500 words for SEO; comprehensive guides at 3,000+ words rank for competitive terms. Write the introduction last — it's easier once you know what you've said. Open loops keep readers scrolling."
	case strings.Contains(lower, "white paper") || strings.Contains(lower, "case study"):
		return "White paper structure: Executive Summary (1 page — the busy reader's version), Problem Statement (backed by data), Solution Framework, Evidence/Case Studies, Implementation Guide, Conclusion + CTA. Case studies: lead with the result ('Company X reduced costs by 40%'), then context, challenge, solution, results with numbers, quote from stakeholder. Concrete numbers trump adjectives every time."
	case strings.Contains(lower, "ghostwrite") || strings.Contains(lower, "editing") || strings.Contains(lower, "proofread"):
		return "Ghostwriting: nail the voice first — read 20+ examples of the subject's writing, identify sentence length patterns, vocabulary choices, and recurring phrases. Editing: read aloud to catch rhythm issues, cut every word that doesn't earn its place, eliminate passive voice where it hides the actor, and ensure each paragraph has one clear point. Proofreading is the last pass — focus only on typos and grammar, not content."
	case strings.Contains(lower, "essay writing") || strings.Contains(lower, "narrative"):
		return "Essay structure: strong thesis in the first paragraph (not 'In this essay I will...'), body paragraphs each with one point + evidence + analysis, and a conclusion that advances beyond the introduction — don't just repeat it. For narrative writing: show don't tell, use scene-setting details selectively (one vivid specific beats five vague generalities), and control pacing — slow down for important moments, skip time for transitions."
	default:
		return "Tell me what you're writing (blog post, sales page, email, essay, script), who the audience is, and the goal (sell, educate, persuade, entertain). Paste any draft you have and I'll give specific, line-level feedback."
	}
}

// ── Design Advisor ────────────────────────────────────────────────────────────

func handleDesign(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "figma") || strings.Contains(lower, "wireframe") || strings.Contains(lower, "mockup"):
		return "Figma workflow: start with low-fidelity wireframes (components off, greyscale) to validate layout before investing in visual design. Use Auto Layout from the start — it makes responsive resizing and handoff clean. Components + variants for all repeated elements. Name layers descriptively; developers will thank you. Prototype with smart animate for smooth interactions. What are you designing — web app, mobile, or marketing site?"
	case strings.Contains(lower, "ui design") || strings.Contains(lower, "user interface"):
		return "UI design principles: 8px grid system for consistent spacing, minimum 4.5:1 contrast ratio (WCAG AA), 16px minimum font size for body text, primary action should be visually dominant (size, colour, position). Reduce cognitive load: one primary action per screen, progressive disclosure for complex forms. The best UI is invisible — users accomplish their goal without noticing the interface."
	case strings.Contains(lower, "ux design") || strings.Contains(lower, "user experience"):
		return "UX research methods by stage: Discovery (user interviews, diary studies) → Definition (affinity mapping, journey maps, Jobs-to-be-Done) → Design (sketching, wireframing, prototyping) → Validation (usability testing, A/B tests, analytics). Five users in a usability test reveal 85% of major issues. Test early with low-fidelity — the higher the fidelity, the less honest feedback you get."
	case strings.Contains(lower, "design system") || strings.Contains(lower, "typography"):
		return "Design system foundation: token layer (colours, spacing, typography as variables) → component library (built on tokens) → pattern library (composed components for common tasks) → documentation. Typography: choose a maximum of 2 typefaces (one for headings, one for body), establish a type scale with a consistent ratio (1.25 or 1.333), and set line-height at 1.5× for body text. Legibility beats creativity in UI typography."
	case strings.Contains(lower, "branding") || strings.Contains(lower, "logo design") || strings.Contains(lower, "color palette"):
		return "Brand colour palette: primary (brand recognition, CTAs), secondary (supporting elements), neutral (backgrounds, text), semantic (success green, error red). Limit to 5–7 colours total. For logo design: works in black and white first — if it fails there, colour won't save it. Test at 16px favicon size and 1m billboard scale. Brand voice and logo must tell the same story. What's the brand personality (e.g., trustworthy, bold, playful)?"
	default:
		return "Describe your design challenge: a specific UI problem, a product to wireframe, a design system to build, or a brand identity to develop. Share the platform (web, iOS, Android), the user, and their goal."
	}
}

// ── Video Strategist ──────────────────────────────────────────────────────────

func handleVideo(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "youtube script") || strings.Contains(lower, "youtube strategy"):
		return "YouTube script structure: Hook (first 30 seconds — state the payoff immediately, not 'in this video I'll show you...'), Context (why this matters), Main Content (deliver the promise, use pattern interrupts every 2–3 minutes), and CTA (subscribe + next video). Retention cliff happens at 30 seconds — spend 80% of your prep there. Algorithm rewards watch time and click-through rate: optimise thumbnail + title first."
	case strings.Contains(lower, "video script") || strings.Contains(lower, "reel script"):
		return "Short-form video (Reels/Shorts/TikTok) script: Hook (first 1–2 seconds — visual or verbal pattern interrupt), Value Delivery (tight, no filler), Loop Trigger (ending that makes viewers rewatch). For educational content: 'Here's what most people get wrong about X...' outperforms listicles. For Reels: caption every word (80% watched muted), vertical 9:16, show your face in the first frame."
	case strings.Contains(lower, "thumbnail design"):
		return "High-CTR thumbnail formula: one human face with strong emotion (curiosity, surprise, shock), large bold text (3–5 words max, readable at mobile size), high-contrast colour block behind text, and a clear visual story that creates an open loop with the title. A/B test thumbnails — YouTube's built-in A/B testing is worth using. Avoid misleading thumbnails — they spike CTR but tank retention, which kills long-term reach."
	case strings.Contains(lower, "video production") || strings.Contains(lower, "filming") || strings.Contains(lower, "b-roll"):
		return "Production basics: lighting is more important than camera — a $200 ring light + diffusion beats a $3,000 camera in bad light. Audio is more important than lighting — a $100 Rode microphone is the single best upgrade for most creators. B-roll: film 3–5 seconds more than you think you need from each angle. Shoot 4K if possible; downscale to 1080p for delivery — gives you crop flexibility in editing."
	case strings.Contains(lower, "premiere pro") || strings.Contains(lower, "final cut") || strings.Contains(lower, "podcast episode"):
		return "Editing workflow: sync audio first, rough cut removing dead air and filler words, fine cut for pacing and story, colour grade (Lumetri in Premiere, built-in in Final Cut), sound mix (dialogue at -12 to -6 dBFS, music at -24 dBFS, SFX at -18 dBFS). For podcasts: record at 44.1kHz/16-bit WAV, use Auphonic for levelling and noise reduction, export at 128kbps AAC for most platforms."
	default:
		return "Tell me what you're creating: YouTube long-form, short-form Reel, a video ad, a podcast, or a documentary? Share your topic, target audience, and channel size or goal. I can help with scripting, structure, thumbnails, or production."
	}
}

// ── Travel Planner ────────────────────────────────────────────────────────────

func handleTravel(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "itinerary"):
		return "Itinerary planning principle: group activities geographically to minimise transit time, not chronologically by interest. Build in one unplanned afternoon per three days — best travel memories are often unscheduled. Balance activity density: 2–3 things per day maximum for sustainable enjoyment. First morning: arrive, orientate, low-key. What destination, how many days, and what's your travel style — cultural, adventure, relaxation, or mixed?"
	case strings.Contains(lower, "visa application"):
		return "Visa checklist: check processing time (apply 6–8 weeks early for visas requiring embassy appointments), gather documents (valid passport >6 months beyond travel dates, bank statements, return tickets, accommodation proof, travel insurance), and check whether an e-visa, visa on arrival, or consulate appointment is required. Double-check requirements on the official embassy website — third-party sites are often outdated. What destination and your passport nationality?"
	case strings.Contains(lower, "flight booking"):
		return "Flight booking hacks: best prices are typically found 6–8 weeks before domestic flights and 2–3 months before international. Google Flights' price calendar and 'Explore' mode find the cheapest dates. Clear your cookies or use incognito (prices don't actually change with cookies, but it's a myth worth eliminating). Set price alerts. Tuesdays and Wednesdays are historically cheapest for departures. Budget airlines: compare total cost with baggage fees."
	case strings.Contains(lower, "budget travel") || strings.Contains(lower, "backpack travel"):
		return "Budget travel fundamentals: accommodation (hostels, Couchsurfing, work exchanges), transport (buses/trains beat flying for routes under 6 hours after airport time), food (eat where locals eat — street food is often safer and better than tourist restaurants), and activities (free walking tours, national parks, public beaches). Southeast Asia and Eastern Europe offer the best value for money globally. What's your budget per day and destination?"
	case strings.Contains(lower, "travel insurance") || strings.Contains(lower, "packing list"):
		return "Travel insurance must-covers: medical evacuation (minimum $250,000 — air ambulances cost $50,000+), trip cancellation/interruption, baggage delay/loss, and 24/7 assistance. Packing list principle: lay everything out, remove 30%, then pack. Roll don't fold. One carry-on is achievable for 2-week trips. Essentials often forgotten: power adaptor, offline maps downloaded, passport copies stored in email and cloud, and a basic first-aid kit."
	default:
		return "Tell me where you're going, for how long, and what matters most to you (culture, food, adventure, cost, comfort). I'll help with itinerary planning, visas, booking strategies, and on-the-ground tips."
	}
}

// ── Mindset Coach ─────────────────────────────────────────────────────────────

func handleMindset(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "growth mindset") || strings.Contains(lower, "limiting belief"):
		return "Growth mindset (Dweck): the core shift is from 'I'm not good at X' to 'I'm not good at X yet.' Identify your specific fixed-mindset triggers — failure, criticism, success of others. Limiting beliefs follow a pattern: an early experience → a conclusion drawn → evidence-seeking that confirms it. Challenge the original conclusion, not the evidence. Write it as: 'I believed X because Y, but the actual lesson was Z.'"
	case strings.Contains(lower, "stoic philosophy") || strings.Contains(lower, "mental model"):
		return "Stoic practices: negative visualisation (premeditatio malorum — imagine the worst outcome calmly, then act) reduces anxiety and builds resilience. The dichotomy of control (Epictetus): focus energy only on what is within your influence. Memento mori — reflect on mortality to clarify priorities. Mental model: 'What would I advise a friend in this situation?' bypasses ego and gives clearer guidance than introspection alone."
	case strings.Contains(lower, "journaling practice") || strings.Contains(lower, "gratitude practice"):
		return "Effective journaling: daily prompts over free-form writing. Morning: 'What would make today a win? What am I grateful for? What am I avoiding?' Evening: 'What did I learn? Where did I fall short? What will I do differently?' Five minutes of structured journaling beats 30 minutes of stream-of-consciousness. Gratitude journaling: specificity matters — 'I'm grateful my partner made coffee this morning' is more effective than 'I'm grateful for my family.'"
	case strings.Contains(lower, "discipline habit") || strings.Contains(lower, "self improvement"):
		return "Habit formation (James Clear): identity before action — 'I am someone who exercises' before 'I will exercise.' Implementation intention: 'I will do X at time Y in location Z' increases follow-through by 2–3×. Habit stacking: attach new habits to existing ones. Friction engineering: make desired habits 20 seconds easier, undesired habits 20 seconds harder. Start with 2-minute versions of big habits — prove the identity first."
	case strings.Contains(lower, "cognitive bias") || strings.Contains(lower, "mindfulness"):
		return "Key cognitive biases in decision-making: confirmation bias (seek disconfirming evidence deliberately), sunk-cost fallacy (evaluate the future path independently of past investment), availability heuristic (vivid recent events distort probability estimates). Mindfulness as a tool: 3 minutes of focused attention on breath before an important decision reduces impulsive choices measurably. The goal is not to eliminate bias but to add a pause between stimulus and response."
	default:
		return "Tell me what you're working on: overcoming a specific belief, building a daily practice, navigating a difficult decision, or developing long-term discipline. I use evidence-based frameworks from cognitive psychology, stoicism, and behavioural science."
	}
}

// ── Food & Nutrition ──────────────────────────────────────────────────────────

func handleFood(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "meal prep") || strings.Contains(lower, "grocery list"):
		return "Meal prep system: cook proteins in bulk (roast chicken thighs, bake salmon, cook legumes), prepare two grain bases (rice, quinoa), and wash/chop vegetables on Sunday. Combine differently each day to avoid repetition. Grocery list principle: shop the perimeter first (produce, protein, dairy), then add staples. Keep a 'pantry inventory' note on your phone — the biggest grocery waste comes from buying duplicates."
	case strings.Contains(lower, "keto recipe"):
		return "Keto macros: 70–75% fat, 20–25% protein, 5% carbs (typically <25g net carbs/day). High-impact keto foods: avocado, eggs, fatty fish (salmon, mackerel), nuts (macadamia, almonds), leafy greens, full-fat dairy. Common mistake: not enough electrolytes in the first 2 weeks — supplement sodium (3–5g/day), potassium (1–3.5g), and magnesium (300–500mg) to avoid the 'keto flu.' What recipe specifically are you looking for?"
	case strings.Contains(lower, "vegan diet") || strings.Contains(lower, "vegetarian"):
		return "Plant-based protein combining: you don't need to combine at every meal (a myth), but over a day ensure variety — legumes, grains, nuts/seeds, soy. Key nutrients to supplement: B12 (essential — not present in plants), Vitamin D, omega-3 (algae-based DHA/EPA vs fish oil), iodine, and iron (pair with Vitamin C for absorption). Tempeh and edamame are the most complete plant proteins. What's your specific dietary question?"
	case strings.Contains(lower, "mediterranean diet"):
		return "Mediterranean diet principles: olive oil as the primary fat (extra-virgin for raw, regular for cooking), fish 2× per week, legumes 3× per week, vegetables at every meal, fruit for dessert, red meat monthly rather than weekly, and red wine in moderation if at all. The most replicated diet-longevity association in research. Simple daily structure: olive oil + vegetables + legume or grain + small protein."
	case strings.Contains(lower, "recipe") || strings.Contains(lower, "cooking") || strings.Contains(lower, "baking") || strings.Contains(lower, "sous vide") || strings.Contains(lower, "ferment"):
		return "Technique matters more than recipes: learn the mother sauces (béchamel, velouté, hollandaise, espagnole, tomato), knife skills, and how to control heat and seasoning — these compound into infinite dishes. Sous vide: protein-specific temperatures (chicken breast 63°C/145°F for 1.5h, steak 54°C/130°F for 1–4h) then hard sear. Fermentation: salt brine (2–3% by weight), submerge vegetables, 1–3 weeks at room temperature. What are you trying to cook?"
	default:
		return "Tell me what you're looking for: a specific recipe, a meal plan for a goal (weight loss, muscle gain, budget), a technique to learn, or a dietary question. Share any constraints (allergies, equipment, time per meal)."
	}
}

// ── Academic Tutor ────────────────────────────────────────────────────────────

func handleTutor(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "math problem") || strings.Contains(lower, "physics concept") || strings.Contains(lower, "chemistry"):
		return "For maths and science problems: share the exact problem statement, what you've already tried, and where you got stuck. I'll explain the concept, walk through the solution step by step, and then give a similar practice problem to solidify understanding. Learning happens at the edge of struggle — if it feels hard, you're in the right zone. What subject and what level (secondary, undergraduate, graduate)?"
	case strings.Contains(lower, "exam prep") || strings.Contains(lower, "study guide"):
		return "Evidence-based study techniques: spaced repetition (Anki) beats rereading by 3–4× for retention. Active recall (test yourself without looking) beats highlighting. Interleaving subjects (switching between topics) beats blocked practice for exam performance. 50-minute focused blocks with 10-minute breaks (Pomodoro). Sleep is when consolidation happens — cramming the night before is counterproductive. What exam and how many days do you have?"
	case strings.Contains(lower, "homework help"):
		return "Share the question and any work you've started. I'll guide you to the answer rather than simply give it — understanding the path is what matters for future problems. Tell me: which subject, what level, and what specifically is confusing you. For writing assignments, share your draft and argument — I'll give structural and line-level feedback."
	case strings.Contains(lower, "history lesson") || strings.Contains(lower, "literature analysis"):
		return "For history: connect events to their causes (political, economic, social, cultural) and to long-term consequences — examiners reward analysis over narrative. Use primary sources to support claims. For literature: theme, structure, language, and context are the four analytical pillars. Close reading: what is this passage doing, not just saying? How does the author achieve the effect? What's your text or historical period?"
	case strings.Contains(lower, "tutoring") || strings.Contains(lower, "learn concept"):
		return "Learning a new concept efficiently: get the overview first (a good YouTube video or textbook chapter introduction) — understand where this fits before diving into details. Then work examples before trying to derive from first principles. Feynman technique: explain the concept in simple language as if to a 12-year-old; gaps in your explanation reveal gaps in your understanding. What concept are you studying?"
	default:
		return "Tell me the subject, level (school, university, professional exam), and what you're working on — a specific problem, a concept you don't understand, or exam preparation. I'll explain clearly and check your understanding."
	}
}

// ── Language Coach ────────────────────────────────────────────────────────────

func handleLanguage(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "translate") || strings.Contains(lower, "translation"):
		return "For translation: share the source text and target language, plus the context (formal document, casual message, marketing copy, technical manual). Context changes word choice significantly — a legal contract and a social media post in Spanish are very different registers. If it's a long document, share a sample and I'll flag any terms where multiple translations are possible and ask for your preference."
	case strings.Contains(lower, "learn spanish") || strings.Contains(lower, "learn french") || strings.Contains(lower, "learn japanese") || strings.Contains(lower, "learn mandarin"):
		return "Language learning path: Pronunciation first (30 days on sounds) → Core vocabulary (top 1,000 words cover 85% of spoken language — use Anki with spaced repetition) → Grammar basics (tenses, sentence structure) → Immersion (podcasts, shows, reading at your level) → Output (speaking/writing practice). Time to conversational: ~600h for Spanish/French (romance languages for English speakers); ~2,200h for Japanese/Mandarin. What's your current level and goal?"
	case strings.Contains(lower, "grammar correction"):
		return "Share the sentence or paragraph you want corrected and I'll fix it and explain each change. For language learning: understanding why something is wrong is more valuable than just getting the correction. I'll note whether each error is a common mistake for speakers of your native language — that pattern recognition speeds up progress. What language and what level are you at?"
	case strings.Contains(lower, "vocabulary") || strings.Contains(lower, "pronunciation"):
		return "Vocabulary acquisition: learn words in context (sentences, not lists), focus on the 1,000 most frequent words first, and use spaced repetition (Anki). Pronunciation: minimal pairs practice (ship vs sheep) targets the sounds your native language doesn't distinguish. Record yourself, compare to a native speaker, and focus on stress and intonation — they affect comprehensibility more than individual phonemes. What language and what specifically are you working on?"
	case strings.Contains(lower, "language learning"):
		return "Fastest path to fluency: massive input at comprehensible level (i+1: slightly above your current level), consistent daily exposure over occasional intensive sessions, and early speaking output (embarrassment is information). Language exchange apps (Tandem, HelloTalk) provide free native speaker conversation. 30 minutes daily beats 3.5 hours once a week for long-term retention. Which language and why are you learning it? That shapes the best approach."
	default:
		return "Tell me which language you're working on, your current level, and your goal — translation, correction, learning from scratch, or preparing for a specific exam or trip. I'll tailor the guidance accordingly."
	}
}

// ── Strategy Consultant ───────────────────────────────────────────────────────

func handleConsulting(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "swot analysis"):
		return "SWOT analysis is most useful when converted into TOWS strategies: SO (use Strengths to exploit Opportunities), ST (use Strengths to mitigate Threats), WO (overcome Weaknesses via Opportunities), WT (minimise both). The common mistake: identifying the SWOT but not generating strategic options from it. Prioritise by impact × actionability. What business or decision is this for?"
	case strings.Contains(lower, "mckinsey framework") || strings.Contains(lower, "bcg matrix") || strings.Contains(lower, "porter five forces"):
		return "Framework selection: Porter's Five Forces (industry attractiveness analysis — useful before market entry), BCG Matrix (portfolio prioritisation: Stars/Cash Cows/Dogs/Question Marks), McKinsey 7S (organisational alignment: Strategy, Structure, Systems, Shared Values, Style, Staff, Skills). MECE principle (Mutually Exclusive, Collectively Exhaustive) underlies all consulting work — structure your problem so issues don't overlap and nothing is missed. What's the business problem you're framing?"
	case strings.Contains(lower, "market entry"):
		return "Market entry framework: Market sizing (TAM/SAM/SOM, bottoms-up) → Competitive dynamics (Porter's Five Forces) → Entry mode (organic, acquisition, partnership, licensing) → Go-to-market (beachhead segment, channel, positioning) → Financial model (unit economics, payback period, break-even). The most common mistake: choosing a market because it's large rather than because you have a defensible right to win. What market and what's your starting advantage?"
	case strings.Contains(lower, "business model canvas") || strings.Contains(lower, "value chain"):
		return "Business Model Canvas nine blocks: Value Proposition → Customer Segments → Channels → Customer Relationships → Revenue Streams → Key Activities → Key Resources → Key Partnerships → Cost Structure. Complete them in that order. Value chain analysis (Porter): map primary activities (inbound logistics, operations, outbound, marketing, service) and support activities (HR, technology, procurement, infrastructure) to find where you create the most value and where you're weakest."
	case strings.Contains(lower, "strategic plan") || strings.Contains(lower, "due diligence"):
		return "Strategic planning structure: Where are we now? (current state assessment) → Where do we want to be? (vision, goals with 3-year horizon) → How do we get there? (strategic initiatives, prioritised by impact/effort) → How do we know we're on track? (KPIs and quarterly reviews). For due diligence: commercial (market size, competitive position, customer concentration), financial (quality of earnings, working capital), legal (IP ownership, liabilities, contracts), and operational (team, systems, dependencies)."
	default:
		return "Describe your strategic challenge: entering a new market, evaluating a business decision, structuring a turnaround, or building a 3-year plan. Share the company context (size, industry, key constraints) and I'll apply the right framework."
	}
}

// ── Medical Information ───────────────────────────────────────────────────────

func handleMedical(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	disclaimer := "\n\nIMPORTANT: This is general health information only — not medical advice. Always consult a qualified doctor or healthcare professional for diagnosis, treatment decisions, and before changing any medication."
	switch {
	case strings.Contains(lower, "drug interaction") || strings.Contains(lower, "medication") || strings.Contains(lower, "dosage"):
		return "Drug interactions can occur via several mechanisms: pharmacokinetic (one drug affects how another is absorbed, metabolised, or eliminated — many involve CYP450 enzymes in the liver) and pharmacodynamic (additive or opposing effects). Common high-risk combinations: blood thinners + NSAIDs, SSRIs + MAOIs, statins + certain antibiotics. Always disclose all medications (including supplements) to your prescribing doctor and pharmacist before starting anything new." + disclaimer
	case strings.Contains(lower, "symptom") || strings.Contains(lower, "diagnosis"):
		return "Symptoms are the body's signals that something needs attention — they rarely map cleanly to a single diagnosis without clinical examination, history, and often tests. Describe your symptoms (location, duration, severity, what makes them better or worse, associated symptoms) to your doctor for accurate assessment. Concerning signs that warrant urgent care: chest pain, sudden severe headache, difficulty breathing, one-sided weakness, or sudden vision changes." + disclaimer
	case strings.Contains(lower, "side effect"):
		return "Side effects vary by medication class and individual. Common ones (fatigue, nausea, headache) often resolve in 1–2 weeks as the body adjusts. Serious side effects requiring immediate medical attention include allergic reactions (rash, swelling, difficulty breathing), severe gastrointestinal bleeding, liver symptoms (jaundice, dark urine), or any neurological changes. Read the patient information leaflet included with your medication and flag anything unusual to your doctor." + disclaimer
	case strings.Contains(lower, "first aid"):
		return "First aid priorities (DRABC): Danger (make the scene safe) → Response (check consciousness) → Airway (tilt head, lift chin) → Breathing (look, listen, feel) → Circulation (chest compressions if not breathing). CPR rate: 100–120 compressions/minute, 2 inches deep, 30:2 ratio with rescue breaths if trained. For bleeding: direct pressure with a clean cloth for 10+ minutes. For burns: cool running water for 20 minutes — do not use ice or butter." + disclaimer
	case strings.Contains(lower, "chronic illness") || strings.Contains(lower, "treatment option"):
		return "Chronic condition management typically involves a combination of: lifestyle modification (diet, exercise, sleep, stress management), pharmacological treatment, monitoring of biomarkers (blood pressure, blood glucose, lipid panels), and regular specialist reviews. Patient self-advocacy is important — ask your doctor about the evidence base for treatment options, the monitoring plan, and what targets you are trying to achieve. Second opinions are appropriate for serious diagnoses." + disclaimer
	default:
		return "Describe your health question: a symptom, a medication, a condition you've been diagnosed with, or general wellness. I provide general health information to help you prepare better questions for your doctor. Always consult a qualified healthcare professional for personal medical decisions." + disclaimer
	}
}

// ── Supply Chain Expert ───────────────────────────────────────────────────────

func handleSupplyChain(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "demand forecast"):
		return "Demand forecasting methods: moving average (simple, smooths noise), exponential smoothing (weights recent data more heavily), ARIMA (captures seasonality and trends), and machine learning (XGBoost, LSTM for complex patterns). Forecast error metrics: MAPE (mean absolute percentage error — target <10% for stable SKUs), bias (systematic over/under-forecasting). Collaborative forecasting (sharing data with suppliers and customers) typically reduces forecast error 20–30%. What's your planning horizon and SKU complexity?"
	case strings.Contains(lower, "logistics") || strings.Contains(lower, "freight") || strings.Contains(lower, "import export"):
		return "Freight mode selection: ocean (cheapest per unit, 15–45 days), air (fastest, 5–10× the cost of ocean, <5 days), rail (mid-cost and time for transcontinental), road (last-mile and regional). Incoterms define who is responsible at each transfer point — EXW (buyer takes all risk) to DDP (seller delivers to buyer's door). For imports: HS codes determine duty rates; a customs broker is worth the cost for anything beyond simple shipments. What origin-destination and what product?"
	case strings.Contains(lower, "lean manufacturing") || strings.Contains(lower, "six sigma"):
		return "Lean manufacturing principles: eliminate the 8 wastes (DOWNTIME: Defects, Overproduction, Waiting, Non-utilised talent, Transport, Inventory, Motion, Extra processing). Value Stream Mapping identifies waste visually. Six Sigma (DMAIC: Define, Measure, Analyse, Improve, Control) drives process quality to 3.4 defects per million opportunities. Lean Six Sigma combines both: Lean for speed, Six Sigma for quality. Start with a pilot value stream rather than a company-wide rollout."
	case strings.Contains(lower, "procurement") || strings.Contains(lower, "vendor management"):
		return "Procurement strategy: segment suppliers by strategic importance × supply risk (Kraljic matrix): Strategic (partner closely), Leverage (drive competition), Bottleneck (secure supply), Non-critical (automate and streamline). Vendor management: SLAs with measurable KPIs, quarterly business reviews for strategic suppliers, dual-sourcing for single points of failure. Total Cost of Ownership (TCO) beats price-only analysis — include quality costs, lead time, and relationship costs."
	case strings.Contains(lower, "warehouse management") || strings.Contains(lower, "supply chain"):
		return "Warehouse layout principles: fast movers near outbound dock (ABC velocity analysis — A items are 20% of SKUs but 80% of picks), dedicated vs shared bin locations, FIFO for perishables (FEFO for expiry dates). KPIs: order fulfilment accuracy (target >99.5%), on-time dispatch, inventory accuracy (cycle count programme). WMS implementation: define processes before configuring the system, not after. What's your current challenge — capacity, accuracy, speed, or cost?"
	default:
		return "Describe your supply chain challenge: demand planning, supplier issues, logistics cost reduction, inventory management, or a disruption you're navigating. Share your industry and scale (number of SKUs, warehouses, annual shipments) for specific recommendations."
	}
}
