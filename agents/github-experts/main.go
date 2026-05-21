// Package main is the reference implementation of the
// `gh.review.code-quality` agent specified in
// genesis/agents-catalog/github-experts.md.
//
// What this binary does:
//   1. Accepts a GitHub PR URL as input.
//   2. Fetches the unified diff via the GitHub REST API (no token needed
//      for public repos at low volume).
//   3. Applies a deterministic rule-set + heuristic checks to the diff.
//   4. Emits a structured review on stdout in the JSON schema documented
//      below.
//
// What this binary does NOT do (yet):
//   - On-chain registration. Registration lands when the agentic-registry
//     CosmWasm contract deploys to Neutron testnet. The structured output
//     is designed so once registration is wired, MsgSubmitResponse just
//     pins this JSON to IPFS and posts the CID.
//   - Call out to an LLM. The whole point of the reference is that it
//     works in CI without secrets. Operators implementing their own
//     version under the same `gh.review.code-quality` catalog ID are
//     free (and encouraged) to use Claude / GPT / Gemini / whatever
//     produces better reviews.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// Output is the JSON envelope this agent emits. Stable across versions —
// changes are versioned via `schema_version`.
type Output struct {
	SchemaVersion string         `json:"schema_version"`
	CatalogID     string         `json:"catalog_id"`
	PR            string         `json:"pr_url"`
	Summary       string         `json:"summary"`
	Findings      []Finding      `json:"findings"`
	Stats         Stats          `json:"stats"`
}

type Finding struct {
	Severity string `json:"severity"` // "info" | "warn" | "error"
	Category string `json:"category"` // "secret" | "test-coverage" | "style" | "logic" | "perf"
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
	Message  string `json:"message"`
}

type Stats struct {
	FilesChanged int `json:"files_changed"`
	LinesAdded   int `json:"lines_added"`
	LinesRemoved int `json:"lines_removed"`
}

func main() {
	prURL := flag.String("pr", "", "Full GitHub PR URL (e.g. https://github.com/sumitkoul23/ai-agentics-bots/pull/7)")
	flag.Parse()

	if *prURL == "" {
		fmt.Fprintln(os.Stderr, "usage: github-experts -pr <pr-url>")
		os.Exit(2)
	}

	diff, err := fetchDiff(*prURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch diff: %v\n", err)
		os.Exit(1)
	}

	out := review(*prURL, diff)
	b, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(b))
}

// fetchDiff turns a PR URL into a unified diff via the .diff suffix that
// GitHub serves. No auth required for public repos.
func fetchDiff(prURL string) (string, error) {
	if !strings.HasPrefix(prURL, "https://github.com/") {
		return "", fmt.Errorf("only https://github.com/ URLs supported")
	}
	url := strings.TrimSuffix(prURL, "/") + ".diff"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "text/plain")
	req.Header.Set("User-Agent", "agentic-gh-experts/0.1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024)) // 5 MB cap
	return string(body), err
}

// review applies the deterministic checks. Each rule is small and
// independently testable.
func review(prURL, diff string) Output {
	out := Output{
		SchemaVersion: "0.1.0",
		CatalogID:     "gh.review.code-quality",
		PR:            prURL,
	}

	out.Stats = countStats(diff)
	out.Findings = append(out.Findings, checkSecrets(diff)...)
	out.Findings = append(out.Findings, checkTestCoverage(diff)...)
	out.Findings = append(out.Findings, checkConsole(diff)...)
	out.Findings = append(out.Findings, checkTODO(diff)...)
	out.Findings = append(out.Findings, checkLargePR(out.Stats)...)

	out.Summary = summarise(out.Findings, out.Stats)
	return out
}

// countStats: lines added / removed / files touched.
func countStats(diff string) Stats {
	var s Stats
	for _, line := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(line, "diff --git "):
			s.FilesChanged++
		case strings.HasPrefix(line, "+++"), strings.HasPrefix(line, "---"):
			continue
		case strings.HasPrefix(line, "+"):
			s.LinesAdded++
		case strings.HasPrefix(line, "-"):
			s.LinesRemoved++
		}
	}
	return s
}

// checkSecrets: flag patterns that look like hardcoded credentials.
// Conservative rule-set; the security-focused catalog entry
// (`gh.review.security`) does the heavy lifting.
var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(api[_-]?key|secret|token|password)\s*[:=]\s*["'][^"']{16,}["']`),
	regexp.MustCompile(`-----BEGIN (?:RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`),
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),       // AWS access key
	regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{36,}`), // GitHub token
	regexp.MustCompile(`sk-[A-Za-z0-9]{32,}`),    // OpenAI / Anthropic style
}

func checkSecrets(diff string) []Finding {
	var out []Finding
	for _, line := range strings.Split(diff, "\n") {
		if !strings.HasPrefix(line, "+") || strings.HasPrefix(line, "+++") {
			continue
		}
		for _, re := range secretPatterns {
			if re.MatchString(line) {
				out = append(out, Finding{
					Severity: "error",
					Category: "secret",
					Message:  "looks like a hardcoded credential — move to env / secrets manager",
				})
				break
			}
		}
	}
	return out
}

// checkTestCoverage: warn if a non-trivial src change has no test diff.
func checkTestCoverage(diff string) []Finding {
	srcChanged := false
	testChanged := false
	for _, line := range strings.Split(diff, "\n") {
		if !strings.HasPrefix(line, "diff --git ") {
			continue
		}
		path := strings.TrimPrefix(line, "diff --git a/")
		if i := strings.Index(path, " "); i > 0 {
			path = path[:i]
		}
		lower := strings.ToLower(path)
		switch {
		case strings.Contains(lower, "test") || strings.HasSuffix(lower, ".test.ts") ||
			strings.HasSuffix(lower, "_test.go") || strings.HasSuffix(lower, "_test.rs"):
			testChanged = true
		case strings.HasSuffix(lower, ".go") || strings.HasSuffix(lower, ".ts") ||
			strings.HasSuffix(lower, ".tsx") || strings.HasSuffix(lower, ".rs") ||
			strings.HasSuffix(lower, ".py"):
			srcChanged = true
		}
	}
	if srcChanged && !testChanged {
		return []Finding{{
			Severity: "warn",
			Category: "test-coverage",
			Message:  "source files changed but no tests added or modified",
		}}
	}
	return nil
}

// checkConsole: flag stray console.log / fmt.Println / print() in non-test files.
var consolePatterns = []*regexp.Regexp{
	regexp.MustCompile(`^\+.*\bconsole\.log\(`),
	regexp.MustCompile(`^\+.*\bfmt\.Println\(`),
	regexp.MustCompile(`^\+.*\bprint\s*\(`),
	regexp.MustCompile(`^\+.*\bdbg!\s*\(`),
}

func checkConsole(diff string) []Finding {
	var out []Finding
	inTest := false
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "diff --git ") {
			lower := strings.ToLower(line)
			inTest = strings.Contains(lower, "test")
			continue
		}
		if inTest {
			continue
		}
		for _, re := range consolePatterns {
			if re.MatchString(line) {
				out = append(out, Finding{
					Severity: "info",
					Category: "style",
					Message:  "debug-print left in non-test code",
				})
				break
			}
		}
	}
	return out
}

// checkTODO: count newly-added TODO / FIXME / HACK comments.
var todoRe = regexp.MustCompile(`(?i)^\+.*\b(TODO|FIXME|HACK|XXX)\b`)

func checkTODO(diff string) []Finding {
	count := 0
	for _, line := range strings.Split(diff, "\n") {
		if todoRe.MatchString(line) {
			count++
		}
	}
	if count > 0 {
		return []Finding{{
			Severity: "info",
			Category: "style",
			Message:  fmt.Sprintf("%d new TODO/FIXME/HACK comment(s) introduced", count),
		}}
	}
	return nil
}

// checkLargePR: a large PR is itself a code-quality risk.
func checkLargePR(s Stats) []Finding {
	if s.LinesAdded+s.LinesRemoved > 1500 {
		return []Finding{{
			Severity: "warn",
			Category: "logic",
			Message:  fmt.Sprintf("very large PR (%d +/-%d lines) — consider splitting", s.LinesAdded, s.LinesRemoved),
		}}
	}
	return nil
}

// summarise produces a 1-line human-readable summary used as the
// MsgSubmitResponse on-chain attribute once registration is wired.
func summarise(findings []Finding, s Stats) string {
	if len(findings) == 0 {
		return fmt.Sprintf("Reviewed %d file(s), +%d/-%d lines. No issues flagged by gh.review.code-quality rule-set.",
			s.FilesChanged, s.LinesAdded, s.LinesRemoved)
	}
	bySev := map[string]int{}
	for _, f := range findings {
		bySev[f.Severity]++
	}
	return fmt.Sprintf("Reviewed %d file(s), +%d/-%d lines. Findings: %d error / %d warn / %d info.",
		s.FilesChanged, s.LinesAdded, s.LinesRemoved,
		bySev["error"], bySev["warn"], bySev["info"])
}
