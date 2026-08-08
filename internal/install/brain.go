package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wassermanproductions/slate-installer/internal/logx"
)

type BrainGuide struct {
	ClaudeInstalled bool   `json:"claudeInstalled"`
	ClaudeLoggedIn  bool   `json:"claudeLoggedIn"`
	CodexInstalled  bool   `json:"codexInstalled"`
	Message         string `json:"message"`
	NextSteps       []string `json:"nextSteps"`
}

// AssessBrain reports brain status without automating OAuth.
func AssessBrain() BrainGuide {
	var g BrainGuide
	claude := which("claude")
	if claude == "" {
		shim := filepath.Join(os.Getenv("APPDATA"), "npm", "claude.cmd")
		if st, err := os.Stat(shim); err == nil && !st.IsDir() {
			claude = shim
		}
	}
	g.ClaudeInstalled = claude != ""
	g.CodexInstalled = which("codex") != ""

	if g.ClaudeInstalled {
		// claude auth status — best effort
		out, err := exec.Command(claude, "auth", "status").CombinedOutput()
		logx.Log("claude auth status: %v %s", err, string(out))
		s := strings.ToLower(string(out))
		g.ClaudeLoggedIn = err == nil && !strings.Contains(s, "not logged") && !strings.Contains(s, "login")
	}

	steps := []string{}
	if !g.ClaudeInstalled {
		steps = append(steps,
			"Optional but recommended: install Claude Code so Slate can help write prompts.",
			"In a terminal run:  npm install -g @anthropic-ai/claude-code",
			"Then run:  claude auth login   (opens your browser — you approve)",
		)
	} else if !g.ClaudeLoggedIn {
		steps = append(steps,
			"Claude Code is installed. Sign in with:  claude auth login",
			"Approve in the browser, then open Slate and pick Brain: Claude Code.",
		)
	} else {
		steps = append(steps, "Claude Code looks ready. In Slate, set Brain to Claude Code and try the Brain pill.")
	}
	steps = append(steps,
		"You can also use a free local model later (Ollama / LM Studio) — no account needed.",
		"Slate never asks for an API key.",
	)
	g.NextSteps = steps
	g.Message = "AI brain setup (optional last step)"
	return g
}

// OfferInstallClaude runs npm install -g only when user confirms (called from UI).
func OfferInstallClaude() (string, error) {
	npm := which("npm.cmd")
	if npm == "" {
		npm = which("npm")
	}
	if npm == "" {
		return "", fmt.Errorf("npm not found — install Node.js first")
	}
	logx.Log("npm install -g @anthropic-ai/claude-code")
	cmd := exec.Command(npm, "install", "-g", "@anthropic-ai/claude-code")
	out, err := cmd.CombinedOutput()
	logx.Log("claude-code install: %s", string(out))
	if err != nil {
		return "", fmt.Errorf("install failed: %v\n%s", err, truncate(string(out), 500))
	}
	return "Claude Code installed. Next: open a terminal and run  claude auth login", nil
}

func which(name string) string {
	p, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return p
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
