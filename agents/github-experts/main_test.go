package main

import (
	"strings"
	"testing"
)

// Sample diff with one new file + a secret + a TODO + a console.log.
const sampleDiff = `diff --git a/src/handler.ts b/src/handler.ts
new file mode 100644
--- /dev/null
+++ b/src/handler.ts
@@ -0,0 +1,8 @@
+// TODO: refactor this
+export const handler = () => {
+  const apiKey = "sk-1234567890abcdefghijklmnop1234567890";
+  console.log("loaded", apiKey);
+  return apiKey;
+};
diff --git a/src/util.go b/src/util.go
--- a/src/util.go
+++ b/src/util.go
@@ -1,2 +1,3 @@
 package util
+func Hello() string { return "hi" }
`

func TestReviewFindsSecret(t *testing.T) {
	out := review("https://github.com/x/y/pull/1", sampleDiff)
	if hasFinding(out.Findings, "secret") == 0 {
		t.Fatalf("expected secret finding, got: %+v", out.Findings)
	}
}

func TestReviewFindsConsole(t *testing.T) {
	out := review("https://github.com/x/y/pull/1", sampleDiff)
	if hasFinding(out.Findings, "style") == 0 {
		t.Fatalf("expected style (console + TODO) findings, got: %+v", out.Findings)
	}
}

func TestReviewWarnsOnMissingTests(t *testing.T) {
	out := review("https://github.com/x/y/pull/1", sampleDiff)
	if hasFinding(out.Findings, "test-coverage") == 0 {
		t.Fatalf("expected test-coverage warning, got: %+v", out.Findings)
	}
}

func TestStatsCount(t *testing.T) {
	out := review("https://github.com/x/y/pull/1", sampleDiff)
	if out.Stats.FilesChanged != 2 {
		t.Errorf("expected 2 files changed, got %d", out.Stats.FilesChanged)
	}
	if out.Stats.LinesAdded < 5 {
		t.Errorf("expected >=5 lines added, got %d", out.Stats.LinesAdded)
	}
}

func TestSummaryContainsStats(t *testing.T) {
	out := review("https://github.com/x/y/pull/1", sampleDiff)
	if !strings.Contains(out.Summary, "file(s)") {
		t.Errorf("summary missing file count: %q", out.Summary)
	}
}

func hasFinding(findings []Finding, category string) int {
	n := 0
	for _, f := range findings {
		if f.Category == category {
			n++
		}
	}
	return n
}
