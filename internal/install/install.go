package install

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/wassermanproductions/slate-installer/internal/installstatus"
	"github.com/wassermanproductions/slate-installer/internal/logx"
	"github.com/wassermanproductions/slate-installer/internal/paths"
)

type ProgressFn func(step, detail string, percent int)

type Result struct {
	InstallDir       string `json:"installDir"`
	ExePath          string `json:"exePath"`
	StartMenuOK      bool   `json:"startMenuOk"`
	DesktopOK        bool   `json:"desktopOk"`
	UninstallOK      bool   `json:"uninstallOk"`
	SmokeOK          bool   `json:"smokeOk"`
	SmokeDetail      string `json:"smokeDetail"`
	FFmpegOK         bool   `json:"ffmpegOk"`
	Summary          string `json:"summary"`
	ProjectsDir      string `json:"projectsDir"`
	ProjectsPreserved bool  `json:"projectsPreserved"`
	IsUpdate         bool   `json:"isUpdate"`
	SourceSHA        string `json:"sourceSha,omitempty"`
}

// InstallOptions configure install/update while protecting user data.
type InstallOptions struct {
	DesktopShortcut bool
	// SourceSHA from GitHub at install time (for update tracking).
	SourceSHA string
	// ReleaseTag optional latest release label.
	ReleaseTag string
	// IsUpdate changes messaging; same safety rules apply.
	IsUpdate bool
}

// Install copies packaged app to per-user Programs\Slate only.
// NEVER touches paths.ProjectsDir() (user project files).
func Install(packagedAppDir string, desktopShortcut bool, progress ProgressFn) (Result, error) {
	return InstallWithOptions(packagedAppDir, InstallOptions{DesktopShortcut: desktopShortcut}, progress)
}

// InstallWithOptions is the full install/update path with manifest + project protection.
func InstallWithOptions(packagedAppDir string, opts InstallOptions, progress ProgressFn) (Result, error) {
	if progress == nil {
		progress = func(string, string, int) {}
	}
	var res Result
	res.ProjectsDir = paths.ProjectsDir()
	res.ProjectsPreserved = true
	res.IsUpdate = opts.IsUpdate
	res.SourceSHA = opts.SourceSHA

	// Absolute safety: never replace install dir if it collides with projects.
	if err := paths.AssertSafeToReplaceInstallDir(); err != nil {
		logx.Log("BLOCKED: %v", err)
		return res, err
	}
	if paths.IsProtectedPath(packagedAppDir) {
		return res, fmt.Errorf("SAFETY STOP: packaged app path is protected — aborting")
	}

	dest := paths.InstallDir()
	if paths.IsProtectedPath(dest) {
		return res, fmt.Errorf("SAFETY STOP: install path is protected — aborting")
	}

	label := "Installing"
	if opts.IsUpdate {
		label = "Updating"
	}
	progress(label, "Copying Slate program files only (your projects are left alone)…", 94)
	logx.Log("%s %s -> %s (projects protected: %s)", label, packagedAppDir, dest, paths.ProjectsDir())

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return res, err
	}
	// Replace app binaries only — never Documents\Slate
	if err := safeRemoveInstallDir(dest); err != nil {
		return res, err
	}
	if err := copyDir(packagedAppDir, dest); err != nil {
		return res, fmt.Errorf("copy app: %w", err)
	}
	res.InstallDir = dest
	res.ExePath = paths.InstalledExe()
	if !fileExists(res.ExePath) {
		// packager may nest differently
		_ = filepath.Walk(dest, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.EqualFold(info.Name(), "Slate.exe") {
				// flatten: if not at root, still OK
				res.ExePath = path
			}
			return nil
		})
	}
	if !fileExists(res.ExePath) {
		return res, fmt.Errorf("Slate.exe missing after install")
	}

	progress("Shortcuts", "Creating Start Menu shortcut…", 96)
	res.StartMenuOK = createShortcut(paths.StartMenuShortcut(), res.ExePath, dest) == nil
	if opts.DesktopShortcut {
		res.DesktopOK = createShortcut(paths.DesktopShortcut(), res.ExePath, dest) == nil
	}
	res.UninstallOK = registerUninstall(res.ExePath, dest) == nil

	progress("Smoke test", "Launching Slate briefly to verify it starts…", 97)
	res.SmokeOK, res.SmokeDetail = smokeTest(res.ExePath)

	// ffmpeg check
	if p, err := exec.LookPath("ffmpeg"); err == nil {
		out, err2 := exec.Command(p, "-version").CombinedOutput()
		res.FFmpegOK = err2 == nil && strings.Contains(string(out), "ffmpeg")
	}

	// Record version for future update checks (inside install dir only).
	_ = installstatus.WriteManifest(installstatus.Manifest{
		SourceRef:   paths.SourceRef,
		SourceSHA:   opts.SourceSHA,
		ReleaseTag:  opts.ReleaseTag,
		ExePath:     res.ExePath,
		SmokeOK:     res.SmokeOK,
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
	})

	lines := []string{}
	if opts.IsUpdate {
		lines = append(lines, "Slate was updated successfully.")
	} else {
		lines = append(lines, "Slate is installed.")
	}
	lines = append(lines, "Location: "+res.ExePath)
	if res.StartMenuOK {
		lines = append(lines, "Start Menu: Slate")
	}
	if res.SmokeOK {
		lines = append(lines, "Startup check: passed")
	} else {
		lines = append(lines, "Startup check: "+res.SmokeDetail)
	}
	lines = append(lines, "Your projects were NOT modified: "+paths.ProjectsDir())
	res.Summary = strings.Join(lines, "\n")
	progress(label, "Complete. Projects folder untouched.", 99)
	return res, nil
}

// safeRemoveInstallDir deletes only the app install directory after safety checks.
func safeRemoveInstallDir(dest string) error {
	if err := paths.AssertSafeToReplaceInstallDir(); err != nil {
		return err
	}
	if paths.IsProtectedPath(dest) {
		return fmt.Errorf("SAFETY STOP: refused to delete protected path %s", dest)
	}
	lower := strings.ToLower(filepath.Clean(dest))
	if strings.Contains(lower, strings.ToLower(filepath.Join("documents", "slate"))) {
		return fmt.Errorf("SAFETY STOP: path looks like projects folder — %s", dest)
	}
	if !fileExists(dest) && !dirExists(dest) {
		return nil
	}
	logx.Log("Removing previous app files only: %s", dest)
	return os.RemoveAll(dest)
}

func dirExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

func smokeTest(exe string) (bool, string) {
	cmd := exec.Command(exe)
	cmd.Dir = filepath.Dir(exe)
	if err := cmd.Start(); err != nil {
		return false, "could not start: " + err.Error()
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			return false, "exited early: " + err.Error()
		}
		return true, "exited cleanly"
	case <-time.After(5 * time.Second):
		// still running = success; kill
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return true, "running after 5s (ok)"
	}
}

func createShortcut(lnkPath, target, workDir string) error {
	_ = os.MkdirAll(filepath.Dir(lnkPath), 0o755)
	// PowerShell WScript.Shell
	ps := fmt.Sprintf(`
$ws = New-Object -ComObject WScript.Shell
$s = $ws.CreateShortcut(%q)
$s.TargetPath = %q
$s.WorkingDirectory = %q
$s.Description = 'Slate — prompt studio for AI filmmaking'
$s.Save()
`, lnkPath, target, workDir)
	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", ps)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logx.Log("shortcut error: %v %s", err, string(out))
		return err
	}
	return nil
}

func registerUninstall(exe, installDir string) error {
	// HKCU uninstall key — no admin
	// DisplayName, UninstallString -> powershell remove dir + shortcuts + reg
	uninstPS := filepath.Join(installDir, "Uninstall-Slate.ps1")
	// IMPORTANT: never remove Documents\Slate (user projects). Only program files + shortcuts.
	script := fmt.Sprintf(`
$ErrorActionPreference = 'SilentlyContinue'
# Uninstall program files ONLY. Do NOT delete user projects under Documents\Slate.
$proj = Join-Path $env:USERPROFILE 'Documents\Slate'
if ((Resolve-Path -LiteralPath %q -ErrorAction SilentlyContinue) -and ((Resolve-Path %q).Path -eq (Resolve-Path $proj -ErrorAction SilentlyContinue).Path)) {
  Write-Error 'Refusing to uninstall: install path matches projects path'
  exit 1
}
Remove-Item -LiteralPath %q -Recurse -Force
Remove-Item -LiteralPath %q -Force
Remove-Item -LiteralPath %q -Force
Remove-Item -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\Slate' -Recurse -Force
`, installDir, installDir, installDir, paths.StartMenuShortcut(), paths.DesktopShortcut())
	if err := os.WriteFile(uninstPS, []byte(script), 0o644); err != nil {
		return err
	}
	uninstCmd := fmt.Sprintf(`powershell -NoProfile -ExecutionPolicy Bypass -File "%s"`, uninstPS)
	ps := fmt.Sprintf(`
New-Item -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\Slate' -Force | Out-Null
Set-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\Slate' -Name 'DisplayName' -Value 'Slate'
Set-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\Slate' -Name 'DisplayVersion' -Value '0.3.2-win'
Set-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\Slate' -Name 'Publisher' -Value 'Slate by Sam Wasserman — Windows Setup: community port'
Set-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\Slate' -Name 'InstallLocation' -Value %q
Set-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\Slate' -Name 'DisplayIcon' -Value %q
Set-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\Slate' -Name 'UninstallString' -Value %q
Set-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\Slate' -Name 'NoModify' -Value 1 -Type DWord
Set-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\Slate' -Name 'NoRepair' -Value 1 -Type DWord
`, installDir, exe, uninstCmd)
	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", ps)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logx.Log("uninstall reg: %v %s", err, string(out))
		return err
	}
	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		_, err = io.Copy(out, in)
		out.Close()
		return err
	})
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}
