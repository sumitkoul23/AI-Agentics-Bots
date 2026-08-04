package main

// MCP (Model Context Protocol) — lightweight tool layer for Bodhi agents.
//
// Enabled via environment variables:
//   MCP_FILESYSTEM=1   read/write local files
//   MCP_SHELL=1        execute shell commands (use with care)
//   MCP_FETCH=1        HTTP fetch / web browsing
//
// Tools are injected into every agent's system prompt and the LLM can
// call them using a simple XML-like syntax:
//
//   <tool name="shell"><cmd>ls -la ~</cmd></tool>
//
// The hub intercepts these tags in the LLM output, executes the tool,
// appends the result, and re-prompts to get the final answer.

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// ── Tool registry ─────────────────────────────────────────────────────────────

type MCPTool struct {
	Name        string
	Description string
	Enabled     func() bool
	Run         func(args string) string
}

var mcpTools []*MCPTool

func init() {
	mcpTools = []*MCPTool{
		{
			Name:        "shell",
			Description: "Execute a shell command and return stdout+stderr. Usage: <tool name=\"shell\"><cmd>your command</cmd></tool>",
			Enabled:     func() bool { return os.Getenv("MCP_SHELL") == "1" },
			Run:         mcpShell,
		},
		{
			Name:        "read_file",
			Description: "Read a local file. Usage: <tool name=\"read_file\"><path>/abs/or/rel/path</path></tool>",
			Enabled:     func() bool { return os.Getenv("MCP_FILESYSTEM") == "1" },
			Run:         mcpReadFile,
		},
		{
			Name:        "write_file",
			Description: "Write text to a local file. Usage: <tool name=\"write_file\"><path>/path/to/file</path><content>text here</content></tool>",
			Enabled:     func() bool { return os.Getenv("MCP_FILESYSTEM") == "1" },
			Run:         mcpWriteFile,
		},
		{
			Name:        "list_dir",
			Description: "List directory contents. Usage: <tool name=\"list_dir\"><path>/path</path></tool>",
			Enabled:     func() bool { return os.Getenv("MCP_FILESYSTEM") == "1" },
			Run:         mcpListDir,
		},
		{
			Name:        "fetch",
			Description: "Fetch a URL and return the response body (first 4 KB). Usage: <tool name=\"fetch\"><url>https://example.com</url></tool>",
			Enabled:     func() bool { return os.Getenv("MCP_FETCH") == "1" },
			Run:         mcpFetch,
		},
	}
}

// EnabledTools returns the subset of tools that are currently active.
func EnabledTools() []*MCPTool {
	var out []*MCPTool
	for _, t := range mcpTools {
		if t.Enabled() {
			out = append(out, t)
		}
	}
	return out
}

// ToolsSystemPrompt returns a block to prepend to agent system prompts.
func ToolsSystemPrompt() string {
	tools := EnabledTools()
	if len(tools) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("\n\n=== TOOLS AVAILABLE ===\n")
	sb.WriteString("You may call any tool below using the exact XML syntax shown.\n")
	sb.WriteString("After calling a tool, wait for <tool_result> before continuing.\n\n")
	for _, t := range tools {
		sb.WriteString("• " + t.Description + "\n")
	}
	sb.WriteString("\nOnly call one tool at a time. Never invent results — always wait for <tool_result>.\n")
	sb.WriteString("======================\n")
	return sb.String()
}

// ── Tool call parser / executor ───────────────────────────────────────────────

var reToolCall = regexp.MustCompile(`(?s)<tool\s+name="([^"]+)">(.*?)</tool>`)
var reTag = regexp.MustCompile(`(?s)<([a-zA-Z_]+)>(.*?)</[a-zA-Z_]+>`)

// ProcessToolCalls scans LLM output, runs any embedded tool calls, and
// returns the augmented text with <tool_result> blocks inserted.
func ProcessToolCalls(text string) (string, bool) {
	match := reToolCall.FindStringSubmatchIndex(text)
	if match == nil {
		return text, false
	}

	toolName := text[match[2]:match[3]]
	body := text[match[4]:match[5]]

	var result string
	found := false
	for _, t := range EnabledTools() {
		if t.Name == toolName {
			result = t.Run(body)
			found = true
			break
		}
	}
	if !found {
		result = fmt.Sprintf("error: unknown tool %q", toolName)
	}

	// Replace the <tool>…</tool> call with the call + result
	replaced := text[:match[0]] +
		text[match[0]:match[1]] +
		"\n<tool_result>\n" + result + "\n</tool_result>\n" +
		text[match[1]:]

	return replaced, true
}

// tagVal extracts the first <tag>value</tag> from body.
func tagVal(body, tag string) string {
	re := regexp.MustCompile(`(?s)<` + tag + `>(.*?)</` + tag + `>`)
	m := re.FindStringSubmatch(body)
	if len(m) < 2 {
		return strings.TrimSpace(body) // fallback: use entire body
	}
	return strings.TrimSpace(m[1])
}

// ── Tool implementations ──────────────────────────────────────────────────────

func mcpShell(body string) string {
	cmdStr := tagVal(body, "cmd")
	if cmdStr == "" {
		return "error: <cmd> is required"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", cmdStr)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", cmdStr)
	}
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	_ = cmd.Run()
	out := buf.String()
	if len(out) > 8192 {
		out = out[:8192] + "\n…(truncated)"
	}
	return out
}

func mcpReadFile(body string) string {
	path := tagVal(body, "path")
	if path == "" {
		return "error: <path> is required"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "error: " + err.Error()
	}
	content := string(data)
	if len(content) > 16384 {
		content = content[:16384] + "\n…(truncated)"
	}
	return content
}

func mcpWriteFile(body string) string {
	path := tagVal(body, "path")
	content := tagVal(body, "content")
	if path == "" {
		return "error: <path> is required"
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "error creating dirs: " + err.Error()
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "error: " + err.Error()
	}
	return fmt.Sprintf("ok: wrote %d bytes to %s", len(content), path)
}

func mcpListDir(body string) string {
	path := tagVal(body, "path")
	if path == "" {
		path = "."
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return "error: " + err.Error()
	}
	var sb strings.Builder
	for _, e := range entries {
		kind := "file"
		if e.IsDir() {
			kind = "dir "
		}
		sb.WriteString(fmt.Sprintf("%s  %s\n", kind, e.Name()))
	}
	return sb.String()
}

func mcpFetch(body string) string {
	rawURL := tagVal(body, "url")
	if rawURL == "" {
		return "error: <url> is required"
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(rawURL)
	if err != nil {
		return "error: " + err.Error()
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return fmt.Sprintf("HTTP %d\n\n%s", resp.StatusCode, string(data))
}

// ── Capabilities summary (for /capabilities endpoint) ────────────────────────

func MCPCapabilities() map[string]interface{} {
	enabled := []string{}
	for _, t := range EnabledTools() {
		enabled = append(enabled, t.Name)
	}
	return map[string]interface{}{
		"mcp_tools": enabled,
		"tunnel":    os.Getenv("TUNNEL"),
		"voice_stt": os.Getenv("WHISPER_BIN") != "",
		"voice_tts": os.Getenv("PIPER_BIN") != "",
		"vision":    os.Getenv("VISION_MODEL") != "",
		"reasoning": os.Getenv("REASONING_MODE"),
	}
}
