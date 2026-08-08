package prereq

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/wassermanproductions/slate-installer/internal/audit"
	"github.com/wassermanproductions/slate-installer/internal/logx"
)

type ProgressFn func(step, detail string, percent int)

// EnsureAll installs missing Node and ffmpeg. Returns refreshed audit-like notes.
func EnsureAll(progress ProgressFn) error {
	if progress == nil {
		progress = func(string, string, int) {}
	}
	report := audit.Run()

	// Refresh PATH from machine+user registry so post-install tools resolve
	refreshPathFromRegistry()

	if !report.NodeOK {
		progress("Node.js", "Installing Node.js LTS (this can take a few minutes)…", 10)
		if err := installNode(progress); err != nil {
			return err
		}
		refreshPathFromRegistry()
		if _, ok := nodeOK(); !ok {
			return fmt.Errorf("Node.js was installed but is not available yet. Close this installer, open a new window, and run Slate Setup again.")
		}
		progress("Node.js", "Node.js is ready.", 35)
	} else {
		progress("Node.js", "Already installed.", 35)
	}

	if !report.FFmpegOK {
		progress("ffmpeg", "Installing ffmpeg…", 40)
		if err := installFFmpeg(progress); err != nil {
			logx.Log("ffmpeg install warning: %v", err)
			progress("ffmpeg", "Could not auto-install ffmpeg. You can install it later; Slate still works without it for basic prompts.", 55)
		} else {
			refreshPathFromRegistry()
			progress("ffmpeg", "ffmpeg is ready.", 55)
		}
	} else {
		progress("ffmpeg", "Already installed.", 55)
	}

	progress("Prerequisites", "Tools check complete.", 60)
	return nil
}

func installNode(progress ProgressFn) error {
	if which("winget") != "" {
		logx.Log("Installing Node via winget")
		progress("Node.js", "Using Windows Package Manager…", 15)
		cmd := exec.Command("winget", "install", "-e", "--id", "OpenJS.NodeJS.LTS",
			"--accept-package-agreements", "--accept-source-agreements", "--disable-interactivity")
		out, err := cmd.CombinedOutput()
		logx.Log("winget node: %s", truncate(string(out), 2000))
		if err == nil {
			time.Sleep(2 * time.Second)
			refreshPathFromRegistry()
			if _, ok := nodeOK(); ok {
				return nil
			}
		}
		logx.Log("winget node failed or not on PATH yet: %v", err)
	}

	// MSI fallback
	progress("Node.js", "Downloading official Node.js installer…", 20)
	msi := filepath.Join(os.TempDir(), "node-lts.msi")
	url := "https://nodejs.org/dist/v22.14.0/node-v22.14.0-x64.msi"
	if err := downloadFile(url, msi, 5*time.Minute); err != nil {
		return fmt.Errorf("download Node.js: %w", err)
	}
	progress("Node.js", "Running Node.js installer (you may see a Windows prompt)…", 25)
	cmd := exec.Command("msiexec", "/i", msi, "/qn", "/norestart")
	out, err := cmd.CombinedOutput()
	logx.Log("msiexec node: %s err=%v", truncate(string(out), 1000), err)
	if err != nil {
		// try interactive
		cmd = exec.Command("msiexec", "/i", msi)
		_ = cmd.Start()
		return fmt.Errorf("please finish the Node.js installer window, then click Retry or run Slate Setup again")
	}
	time.Sleep(3 * time.Second)
	refreshPathFromRegistry()
	return nil
}

func installFFmpeg(progress ProgressFn) error {
	if which("winget") == "" {
		return fmt.Errorf("winget not available")
	}
	logx.Log("Installing ffmpeg via winget")
	cmd := exec.Command("winget", "install", "-e", "--id", "Gyan.FFmpeg",
		"--accept-package-agreements", "--accept-source-agreements", "--disable-interactivity")
	out, err := cmd.CombinedOutput()
	logx.Log("winget ffmpeg: %s", truncate(string(out), 2000))
	if err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	return nil
}

func nodeOK() (string, bool) {
	p := which("node")
	if p == "" {
		// common install path
		cand := `C:\Program Files\nodejs\node.exe`
		if st, err := os.Stat(cand); err == nil && !st.IsDir() {
			p = cand
		}
	}
	if p == "" {
		return "", false
	}
	out, err := exec.Command(p, "-v").CombinedOutput()
	if err != nil {
		return "", false
	}
	v := strings.TrimSpace(string(out))
	vnum := strings.TrimPrefix(v, "v")
	maj := 0
	fmt.Sscanf(vnum, "%d", &maj)
	return v, maj >= 20
}

func which(name string) string {
	p, err := exec.LookPath(name)
	if err != nil {
		p, err = exec.LookPath(name + ".exe")
		if err != nil {
			return ""
		}
	}
	return p
}

func refreshPathFromRegistry() {
	// Merge Machine + User PATH into process env
	machine := readRegPATH(`HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`)
	user := readRegPATH(`HKCU\Environment`)
	extra := []string{
		`C:\Program Files\nodejs`,
		filepath.Join(os.Getenv("APPDATA"), "npm"),
		`C:\ffmpeg\bin`,
	}
	// Gyan winget often installs under LocalAppData
	if la := os.Getenv("LOCALAPPDATA"); la != "" {
		// search common winget package dirs is hard; append LocalAppData\Microsoft\WinGet\Links
		extra = append(extra, filepath.Join(la, "Microsoft", "WinGet", "Links"))
	}
	parts := []string{}
	seen := map[string]bool{}
	for _, chunk := range []string{os.Getenv("PATH"), user, machine, strings.Join(extra, ";")} {
		for _, p := range strings.Split(chunk, ";") {
			p = strings.TrimSpace(p)
			if p == "" || seen[strings.ToLower(p)] {
				continue
			}
			seen[strings.ToLower(p)] = true
			parts = append(parts, p)
		}
	}
	_ = os.Setenv("PATH", strings.Join(parts, ";"))
	logx.Log("PATH refreshed (%d entries)", len(parts))
}

func readRegPATH(key string) string {
	// use reg query
	// key format HKCU\Environment
	parts := strings.SplitN(key, `\`, 2)
	if len(parts) != 2 {
		return ""
	}
	cmd := exec.Command("reg", "query", parts[0]+`\`+parts[1], "/v", "Path")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	// REG_EXPAND_SZ    Path    value
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Path") && (strings.Contains(line, "REG_")) {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				// last field may be incomplete if spaces — take after type
				idx := strings.Index(line, "REG_")
				if idx >= 0 {
					rest := line[idx:]
					sp := strings.Index(rest, " ")
					if sp >= 0 {
						return strings.TrimSpace(rest[sp+1:])
					}
				}
			}
		}
	}
	return ""
}

func downloadFile(url, dest string, timeout time.Duration) error {
	logx.Log("Download %s -> %s", url, dest)
	// use curl if available (Windows 10+), else powershell
	if which("curl") != "" {
		cmd := exec.Command("curl", "-fsSL", "--connect-timeout", "30", "--max-time", fmt.Sprintf("%d", int(timeout.Seconds())), "-o", dest, url)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("curl: %v %s", err, truncate(string(out), 400))
		}
		return nil
	}
	ps := fmt.Sprintf(`$ProgressPreference='SilentlyContinue'; Invoke-WebRequest -Uri %q -OutFile %q -UseBasicParsing -TimeoutSec %d`, url, dest, int(timeout.Seconds()))
	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", ps)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("download: %v %s", err, truncate(string(out), 400))
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
