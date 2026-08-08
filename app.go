package main

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wassermanproductions/slate-installer/internal/audit"
	"github.com/wassermanproductions/slate-installer/internal/buildapp"
	"github.com/wassermanproductions/slate-installer/internal/install"
	"github.com/wassermanproductions/slate-installer/internal/installstatus"
	"github.com/wassermanproductions/slate-installer/internal/logx"
	"github.com/wassermanproductions/slate-installer/internal/patch"
	"github.com/wassermanproductions/slate-installer/internal/paths"
	"github.com/wassermanproductions/slate-installer/internal/prereq"
	"github.com/wassermanproductions/slate-installer/internal/source"
	"github.com/wassermanproductions/slate-installer/internal/update"
)

// App is the installer host bound to the wizard UI.
type App struct {
	ctx       context.Context
	mu        sync.Mutex
	running   bool
	lastRes   install.Result
	lastGuide install.BrainGuide
	lastSHA   string
	lastTag   string
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	_ = logx.Init()
	logx.Log("Slate Setup started (v%s)", paths.InstallerVersion)
}

func (a *App) emitProgress(step, detail string, percent int) {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, "install:progress", map[string]interface{}{
		"step": step, "detail": detail, "percent": percent,
	})
}

// RunAudit performs pre-flight checks (no changes to the system).
func (a *App) RunAudit() audit.Report {
	return audit.Run()
}

// RunDeepAudit is the stand-alone Audit tool: install health + updates + projects (read-only).
func (a *App) RunDeepAudit() audit.DeepReport {
	return audit.RunDeep()
}

// GetInstallStatus reports whether Slate is already installed successfully.
func (a *App) GetInstallStatus() installstatus.Status {
	return installstatus.Detect()
}

// CheckForUpdates compares installed source SHA to GitHub main.
func (a *App) CheckForUpdates() update.CheckResult {
	r := update.Check()
	if r.LatestSHA != "" {
		a.lastSHA = r.LatestSHA
	}
	if r.LatestReleaseTag != "" {
		a.lastTag = r.LatestReleaseTag
	}
	return r
}

// LogPath returns the install log path for support.
func (a *App) LogPath() string {
	return logx.Path()
}

// OpenLogFolder opens the log file location.
func (a *App) OpenLogFolder() {
	runtime.BrowserOpenURL(a.ctx, "file:///"+filepath.ToSlash(filepath.Dir(paths.TempLog())))
}

// OpenProjectsFolder opens the user projects directory (never modified by Setup).
func (a *App) OpenProjectsFolder() {
	p := paths.ProjectsDir()
	runtime.BrowserOpenURL(a.ctx, "file:///"+filepath.ToSlash(p))
}

// GetPaths exposes install destinations for the UI.
func (a *App) GetPaths() map[string]string {
	return map[string]string{
		"installDir":       paths.InstallDir(),
		"workDir":          paths.WorkDir(),
		"logPath":          paths.TempLog(),
		"sourceRef":        paths.SourceRef,
		"projectsDir":      paths.ProjectsDir(),
		"installerVersion": paths.InstallerVersion,
	}
}

// StartInstall runs a fresh install or repair. desktopShortcut chooses Desktop icon.
func (a *App) StartInstall(desktopShortcut bool) (map[string]interface{}, error) {
	return a.runPipeline(desktopShortcut, false)
}

// StartUpdate forces a fresh source download and reinstall of program files only.
// User project files under Documents\Slate are never touched.
func (a *App) StartUpdate(desktopShortcut bool) (map[string]interface{}, error) {
	return a.runPipeline(desktopShortcut, true)
}

func (a *App) runPipeline(desktopShortcut bool, isUpdate bool) (map[string]interface{}, error) {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return nil, fmt.Errorf("install already running")
	}
	a.running = true
	a.mu.Unlock()
	defer func() {
		a.mu.Lock()
		a.running = false
		a.mu.Unlock()
	}()

	progress := func(step, detail string, percent int) {
		logx.Log("[%d%%] %s — %s", percent, step, detail)
		a.emitProgress(step, detail, percent)
	}

	progress("Safety", "Confirming your project files will not be touched…", 1)
	if err := paths.AssertSafeToReplaceInstallDir(); err != nil {
		return nil, err
	}
	logx.Log("Projects protected at %s", paths.ProjectsDir())

	// Capture latest SHA for manifest (best-effort before build)
	upd := update.Check()
	if upd.LatestSHA != "" {
		a.lastSHA = upd.LatestSHA
	}
	if upd.LatestReleaseTag != "" {
		a.lastTag = upd.LatestReleaseTag
	}

	progress("Check", "Checking your computer…", 2)
	rep := audit.Run()
	if !rep.CanProceed {
		return nil, fmt.Errorf("%s", rep.Summary)
	}

	progress("Tools", "Installing any missing tools…", 8)
	if err := prereq.EnsureAll(progress); err != nil {
		return nil, err
	}

	// Updates always re-download source; first install reuses cache if present.
	forceSource := isUpdate
	src, err := source.Ensure(progress, forceSource)
	if err != nil {
		return nil, err
	}

	progress("Windows fix", "Applying Windows compatibility fix…", 71)
	patchMsg, err := patch.ApplyBrainWindowsPatch(src)
	if err != nil {
		return nil, fmt.Errorf("Windows patch failed: %w", err)
	}
	logx.Log("%s", patchMsg)
	progress("Windows fix", patchMsg, 71)

	appDir, err := buildapp.Build(src, progress)
	if err != nil {
		return nil, err
	}

	res, err := install.InstallWithOptions(appDir, install.InstallOptions{
		DesktopShortcut: desktopShortcut,
		SourceSHA:       a.lastSHA,
		ReleaseTag:      a.lastTag,
		IsUpdate:        isUpdate,
	}, progress)
	if err != nil {
		return nil, err
	}
	a.lastRes = res

	guide := install.AssessBrain()
	a.lastGuide = guide
	progress("Done", "Finished. Projects folder was not modified.", 100)

	return map[string]interface{}{
		"result":           res,
		"brain":            guide,
		"patch":            patchMsg,
		"log":              logx.Path(),
		"projectsDir":      paths.ProjectsDir(),
		"projectsPreserved": true,
		"isUpdate":         isUpdate,
	}, nil
}

// InstallClaudeCode is opt-in only (user clicked the button).
func (a *App) InstallClaudeCode() (string, error) {
	return install.OfferInstallClaude()
}

// GetBrainGuide re-assesses brain status.
func (a *App) GetBrainGuide() install.BrainGuide {
	return install.AssessBrain()
}

// LaunchSlate opens the install folder in Explorer (fallback).
func (a *App) LaunchSlate() {
	exe := a.lastRes.ExePath
	if exe == "" {
		exe = paths.InstalledExe()
	}
	runtime.BrowserOpenURL(a.ctx, "file:///"+filepath.ToSlash(filepath.Dir(exe)))
}

// LaunchSlateProcess actually starts the process.
func (a *App) LaunchSlateProcess() error {
	exe := a.lastRes.ExePath
	if exe == "" {
		exe = paths.InstalledExe()
	}
	return openExe(exe)
}

// OpenExternal opens a URL.
func (a *App) OpenExternal(url string) {
	runtime.BrowserOpenURL(a.ctx, url)
}
