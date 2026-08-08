package audit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/wassermanproductions/slate-installer/internal/logx"
	"golang.org/x/sys/windows"
)

type Check struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	OK       bool   `json:"ok"`
	Required bool   `json:"required"`
	Detail   string `json:"detail"`
	Action   string `json:"action"` // what installer will do if !OK
}

type Report struct {
	Checks        []Check `json:"checks"`
	CanProceed    bool    `json:"canProceed"`
	Summary       string  `json:"summary"`
	WindowsOK     bool    `json:"windowsOk"`
	DiskOK        bool    `json:"diskOk"`
	NodeOK        bool    `json:"nodeOk"`
	FFmpegOK      bool    `json:"ffmpegOk"`
	GitOK         bool    `json:"gitOk"`
	WingetOK      bool    `json:"wingetOk"`
	ClaudeOK      bool    `json:"claudeOk"`
	CodexOK       bool    `json:"codexOk"`
	AlreadyInstalled bool `json:"alreadyInstalled"`
	InstallPath   string  `json:"installPath"`
}

func Run() Report {
	logx.Log("Starting system audit")
	var r Report
	r.Checks = []Check{}

	// Windows version
	ver, winOK := windowsVersion()
	r.WindowsOK = winOK
	r.Checks = append(r.Checks, Check{
		ID: "windows", Label: "Windows version", OK: winOK, Required: true,
		Detail: ver,
		Action: ternary(winOK, "Ready", "Windows 10 or 11 is required"),
	})

	// Disk space
	freeGB, diskOK := freeSpaceGB()
	r.DiskOK = diskOK
	r.Checks = append(r.Checks, Check{
		ID: "disk", Label: "Free disk space", OK: diskOK, Required: true,
		Detail: fmt.Sprintf("%.1f GB free (need about 2 GB)", freeGB),
		Action: ternary(diskOK, "Ready", "Free up disk space, then try again"),
	})

	// winget
	winget := which("winget")
	r.WingetOK = winget != ""
	r.Checks = append(r.Checks, Check{
		ID: "winget", Label: "Windows Package Manager (winget)", OK: r.WingetOK, Required: false,
		Detail: ternary(r.WingetOK, winget, "Not found — Node can still be installed via official MSI"),
		Action: ternary(r.WingetOK, "Will use winget for Node/ffmpeg when needed", "Will download Node MSI if Node is missing"),
	})

	// Node
	nodeV, nodeOK := nodeVersion()
	r.NodeOK = nodeOK
	r.Checks = append(r.Checks, Check{
		ID: "node", Label: "Node.js 20+", OK: nodeOK, Required: true,
		Detail: ternary(nodeOK, "Found: "+nodeV, "Not found or too old"),
		Action: ternary(nodeOK, "Ready", "Installer will install Node.js LTS automatically"),
	})

	// ffmpeg
	ff := which("ffmpeg")
	r.FFmpegOK = ff != ""
	r.Checks = append(r.Checks, Check{
		ID: "ffmpeg", Label: "ffmpeg (for video/audio features)", OK: r.FFmpegOK, Required: false,
		Detail: ternary(r.FFmpegOK, ff, "Not on PATH"),
		Action: ternary(r.FFmpegOK, "Ready", "Installer will install ffmpeg via winget if available"),
	})

	// git optional
	git := which("git")
	r.GitOK = git != ""
	r.Checks = append(r.Checks, Check{
		ID: "git", Label: "Git (optional)", OK: r.GitOK, Required: false,
		Detail: ternary(r.GitOK, git, "Not required — installer downloads a zip instead"),
		Action: "No action needed",
	})

	// brains
	claude := which("claude")
	if claude == "" {
		// npm global shim
		if p := filepath.Join(os.Getenv("APPDATA"), "npm", "claude.cmd"); fileExists(p) {
			claude = p
		}
	}
	r.ClaudeOK = claude != ""
	r.Checks = append(r.Checks, Check{
		ID: "claude", Label: "Claude Code (AI brain)", OK: r.ClaudeOK, Required: false,
		Detail: ternary(r.ClaudeOK, claude, "Not installed — optional, can sign in later"),
		Action: ternary(r.ClaudeOK, "Ready for Claude brain", "You can connect Claude after install (recommended)"),
	})

	codex := which("codex")
	r.CodexOK = codex != ""
	r.Checks = append(r.Checks, Check{
		ID: "codex", Label: "Codex CLI (optional brain)", OK: r.CodexOK, Required: false,
		Detail: ternary(r.CodexOK, codex, "Not installed"),
		Action: "Optional",
	})

	// already installed?
	installExe := filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Slate", "Slate.exe")
	r.AlreadyInstalled = fileExists(installExe)
	r.InstallPath = installExe
	r.Checks = append(r.Checks, Check{
		ID: "existing", Label: "Existing Slate install", OK: true, Required: false,
		Detail: ternary(r.AlreadyInstalled, "Found at "+installExe+" (repair/reinstall OK)", "Not installed yet"),
		Action: ternary(r.AlreadyInstalled, "Installer will repair/update", "Fresh install"),
	})

	r.CanProceed = r.WindowsOK && r.DiskOK
	if !r.CanProceed {
		r.Summary = "This PC cannot install Slate yet. Fix the red items above, then click Check again."
	} else if r.NodeOK && r.FFmpegOK {
		r.Summary = "Looking good! Click Continue to install Slate. A few downloads may still run."
	} else {
		r.Summary = "You can continue. The installer will automatically install missing tools for you."
	}
	logx.Log("Audit complete: canProceed=%v node=%v ffmpeg=%v", r.CanProceed, r.NodeOK, r.FFmpegOK)
	return r
}

func windowsVersion() (string, bool) {
	if runtime.GOOS != "windows" {
		return "Not Windows", false
	}
	maj, min, build := windows.RtlGetNtVersionNumbers()
	// Windows 10+ share major 10; build >= 22000 is Win11
	label := fmt.Sprintf("Windows %d.%d (build %d)", maj, min, build)
	if maj < 10 {
		return label + " — too old", false
	}
	if build >= 22000 {
		label = fmt.Sprintf("Windows 11 (build %d)", build)
	} else {
		label = fmt.Sprintf("Windows 10 (build %d)", build)
	}
	return label, true
}

func freeSpaceGB() (float64, bool) {
	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64
	path, _ := syscall.UTF16PtrFromString("C:\\")
	err := windows.GetDiskFreeSpaceEx(path, &freeBytesAvailable, &totalNumberOfBytes, &totalNumberOfFreeBytes)
	if err != nil {
		// try LOCALAPPDATA drive
		return 10, true
	}
	gb := float64(freeBytesAvailable) / (1024 * 1024 * 1024)
	return gb, gb >= 1.5
}

func which(name string) string {
	p, err := exec.LookPath(name)
	if err != nil {
		p, err = exec.LookPath(name + ".exe")
		if err != nil {
			p, err = exec.LookPath(name + ".cmd")
			if err != nil {
				return ""
			}
		}
	}
	return p
}

func nodeVersion() (string, bool) {
	p := which("node")
	if p == "" {
		return "", false
	}
	out, err := exec.Command(p, "-v").CombinedOutput()
	if err != nil {
		return "", false
	}
	v := strings.TrimSpace(string(out))
	// v20.x.x
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	if len(parts) == 0 {
		return v, false
	}
	maj, _ := strconv.Atoi(parts[0])
	return "v" + v, maj >= 20
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

func ternary(c bool, a, b string) string {
	if c {
		return a
	}
	return b
}

