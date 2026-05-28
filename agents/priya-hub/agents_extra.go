package main

// ── Extra specialist agents (23) ──────────────────────────────────────────────

func taxAgent() *Agent {
	return &Agent{
		ID:   "tax",
		Name: "Tax Strategist",
		Desc: "Tax planning, deductions, IRS/HMRC, capital gains, self-employed tax strategies",
		Keywords: []string{
			"tax", "taxes", "irs", "hmrc", "tax return", "deduction", "write-off",
			"capital gains", "tax loss harvest", "s corp", "llc tax", "self-employed tax",
		},
		Handle: handleTax,
	}
}

func realEstateAgent() *Agent {
	return &Agent{
		ID:   "real-estate",
		Name: "Real Estate Advisor",
		Desc: "Property investment, mortgages, rental income, REITs, house hacking strategies",
		Keywords: []string{
			"real estate", "property invest", "mortgage", "rent vs buy", "reit",
			"rental income", "house hack", "cap rate", "noi", "1031 exchange", "airbnb invest",
		},
		Handle: handleRealEstate,
	}
}

func startupAgent() *Agent {
	return &Agent{
		ID:   "startup",
		Name: "Startup Coach",
		Desc: "Pitch decks, fundraising, cap tables, MVPs, product-market fit, accelerators",
		Keywords: []string{
			"startup", "pitch deck", "fundraise", "raise capital", "seed round",
			"series a", "venture capital", "vc funding", "mvp", "product market fit",
			"term sheet", "cap table", "accelerator",
		},
		Handle: handleStartup,
	}
}

func salesAgent() *Agent {
	return &Agent{
		ID:   "sales",
		Name: "Sales Coach",
		Desc: "Sales strategy, cold calling, objection handling, CRM pipelines, B2B/SaaS sales",
		Keywords: []string{
			"sales strategy", "cold call", "sales pitch", "close the deal",
			"discovery call", "objection handling", "crm pipeline",
			"b2b sales", "saas sales", "quota",
		},
		Handle: handleSales,
	}
}

func marketingAgent() *Agent {
	return &Agent{
		ID:   "marketing",
		Name: "Marketing Strategist",
		Desc: "Growth hacking, funnels, CAC/LTV, SEO, SEM, email campaigns, brand awareness",
		Keywords: []string{
			"marketing strategy", "growth hack", "sales funnel", "lead generation",
			"brand awareness", "customer acquisition", "cac ltv", "a/b test",
			"seo strategy", "sem campaign", "email campaign",
		},
		Handle: handleMarketing,
	}
}

func legalAgent() *Agent {
	return &Agent{
		ID:   "legal",
		Name: "Legal Advisor",
		Desc: "Contracts, NDAs, IP, trademark, patent, incorporation, GDPR, compliance",
		Keywords: []string{
			"contract", "legal advice", "terms of service", "privacy policy", "nda",
			"intellectual property", "trademark", "patent", "copyright",
			"incorporation", "compliance", "gdpr",
		},
		Handle: handleLegal,
	}
}

func hrAgent() *Agent {
	return &Agent{
		ID:   "hr",
		Name: "HR & People",
		Desc: "Hiring, job descriptions, onboarding, performance reviews, equity plans, culture",
		Keywords: []string{
			"hiring plan", "job description", "interview question", "onboard employee",
			"performance review", "compensation", "equity plan", "team culture",
			"remote team", "fire employee",
		},
		Handle: handleHR,
	}
}

func ecommerceAgent() *Agent {
	return &Agent{
		ID:   "ecommerce",
		Name: "E-Commerce Expert",
		Desc: "Shopify, Amazon FBA, dropshipping, product listings, fulfillment, inventory",
		Keywords: []string{
			"shopify", "amazon fba", "dropshipping", "ecommerce store", "product listing",
			"etsy", "woocommerce", "product sourcing", "fulfillment", "inventory",
		},
		Handle: handleEcommerce,
	}
}

func devopsAgent() *Agent {
	return &Agent{
		ID:   "devops",
		Name: "DevOps Engineer",
		Desc: "Infrastructure, Terraform, Helm, CI/CD, cloud architecture, SRE, monitoring",
		Keywords: []string{
			"devops", "infrastructure", "terraform", "helm chart", "container orchestration",
			"cloud architecture", "ansible", "jenkins", "github actions pipeline",
			"site reliability", "sre", "prometheus", "grafana",
		},
		Handle: handleDevops,
	}
}

func dataAgent() *Agent {
	return &Agent{
		ID:   "data",
		Name: "Data Scientist",
		Desc: "ML models, data pipelines, ETL, Pandas/NumPy, BigQuery, Snowflake, dashboards",
		Keywords: []string{
			"data science", "machine learning model", "pandas", "numpy", "jupyter",
			"data pipeline", "etl pipeline", "data warehouse", "bigquery",
			"snowflake", "looker", "tableau", "analytics dashboard",
		},
		Handle: handleData,
	}
}

func securityAgent() *Agent {
	return &Agent{
		ID:   "security",
		Name: "Security Expert",
		Desc: "Cybersecurity, pen testing, OWASP, threat modelling, incident response",
		Keywords: []string{
			"cybersecurity", "penetration test", "pentest", "vulnerability scan",
			"owasp", "sql injection attack", "xss attack", "threat model",
			"security audit", "red team", "incident response",
		},
		Handle: handleSecurity,
	}
}

func web3Agent() *Agent {
	return &Agent{
		ID:   "web3",
		Name: "Web3 Developer",
		Desc: "Smart contracts, Solidity, DeFi, NFTs, DAOs, L2s, ZK proofs",
		Keywords: []string{
			"smart contract", "solidity", "defi protocol", "nft project", "dao governance",
			"tokenomics", "hardhat", "foundry framework", "layer 2",
			"zk proof", "polygon network", "arbitrum",
		},
		Handle: handleWeb3,
	}
}

func writingAgent() *Agent {
	return &Agent{
		ID:   "writing",
		Name: "Writing Coach",
		Desc: "Copywriting, sales copy, blog posts, white papers, ghostwriting, proofreading",
		Keywords: []string{
			"copywriting", "sales copy", "ad copy", "blog post writing", "essay writing",
			"white paper", "case study", "ghostwrite", "editing", "proofread",
			"narrative", "long-form",
		},
		Handle: handleWriting,
	}
}

func designAgent() *Agent {
	return &Agent{
		ID:   "design",
		Name: "Design Advisor",
		Desc: "UI/UX, Figma, wireframes, design systems, branding, typography",
		Keywords: []string{
			"ui design", "ux design", "figma", "wireframe", "mockup",
			"design system", "color palette", "typography", "user interface",
			"user experience", "branding", "logo design",
		},
		Handle: handleDesign,
	}
}

func videoAgent() *Agent {
	return &Agent{
		ID:   "video",
		Name: "Video Strategist",
		Desc: "YouTube scripts, video production, thumbnails, reels, podcast episodes",
		Keywords: []string{
			"youtube script", "video script", "video production", "filming",
			"premiere pro", "final cut", "b-roll", "thumbnail design",
			"youtube strategy", "reel script", "podcast episode",
		},
		Handle: handleVideo,
	}
}

func travelAgent() *Agent {
	return &Agent{
		ID:   "travel",
		Name: "Travel Planner",
		Desc: "Itineraries, visas, flights, hotels, budget travel, travel insurance",
		Keywords: []string{
			"travel plan", "itinerary", "visa application", "flight booking",
			"hotel booking", "backpack travel", "budget travel",
			"travel insurance", "packing list", "travel hack",
		},
		Handle: handleTravel,
	}
}

func mindsetAgent() *Agent {
	return &Agent{
		ID:   "mindset",
		Name: "Mindset Coach",
		Desc: "Growth mindset, mental models, stoicism, journaling, discipline, self-improvement",
		Keywords: []string{
			"growth mindset", "mental model", "cognitive bias", "limiting belief",
			"stoic philosophy", "journaling practice", "gratitude practice",
			"mindfulness", "discipline habit", "self improvement",
		},
		Handle: handleMindset,
	}
}

func foodAgent() *Agent {
	return &Agent{
		ID:   "food",
		Name: "Food & Nutrition",
		Desc: "Recipes, meal prep, grocery lists, special diets, cooking techniques",
		Keywords: []string{
			"recipe", "cooking", "meal prep", "grocery list", "vegetarian",
			"vegan diet", "keto recipe", "mediterranean diet", "sous vide",
			"ferment", "baking", "restaurant",
		},
		Handle: handleFood,
	}
}

func tutorAgent() *Agent {
	return &Agent{
		ID:   "tutor",
		Name: "Academic Tutor",
		Desc: "Homework help, exam prep, study guides, maths, sciences, humanities",
		Keywords: []string{
			"homework help", "exam prep", "study guide", "math problem",
			"physics concept", "chemistry", "history lesson",
			"literature analysis", "tutoring", "learn concept",
		},
		Handle: handleTutor,
	}
}

func languageAgent() *Agent {
	return &Agent{
		ID:   "language",
		Name: "Language Coach",
		Desc: "Translation, language learning, grammar, vocabulary, pronunciation",
		Keywords: []string{
			"translate", "translation", "learn spanish", "learn french",
			"learn japanese", "learn mandarin", "grammar correction",
			"vocabulary", "language learning", "pronunciation",
		},
		Handle: handleLanguage,
	}
}

func consultingAgent() *Agent {
	return &Agent{
		ID:   "consulting",
		Name: "Strategy Consultant",
		Desc: "Business strategy, SWOT, market entry, McKinsey frameworks, business model canvas",
		Keywords: []string{
			"business strategy", "strategic plan", "swot analysis", "market entry",
			"mckinsey framework", "bcg matrix", "porter five forces",
			"value chain", "business model canvas", "due diligence",
		},
		Handle: handleConsulting,
	}
}

func medicalAgent() *Agent {
	return &Agent{
		ID:   "medical",
		Name: "Medical Information",
		Desc: "General health information on symptoms, medications, conditions — always consult a doctor",
		Keywords: []string{
			"symptom", "diagnosis", "medication", "drug interaction", "dosage",
			"side effect", "medical condition", "prescription",
			"treatment option", "chronic illness", "first aid",
		},
		Handle: handleMedical,
	}
}

func supplyChainAgent() *Agent {
	return &Agent{
		ID:   "supply-chain",
		Name: "Supply Chain Expert",
		Desc: "Logistics, procurement, freight, demand forecasting, lean, Six Sigma, import/export",
		Keywords: []string{
			"supply chain", "logistics", "procurement", "freight", "vendor management",
			"demand forecast", "lean manufacturing", "six sigma",
			"warehouse management", "import export",
		},
		Handle: handleSupplyChain,
	}
}
