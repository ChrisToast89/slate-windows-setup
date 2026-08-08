package buildapp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/wassermanproductions/slate-installer/internal/logx"
)

type ProgressFn func(step, detail string, percent int)

// Build runs npm ci, verifies Electron binary, npm run build, then packages win32-x64.
// Returns path to packaged app directory containing Slate.exe.
func Build(sourceRoot string, progress ProgressFn) (string, error) {
	if progress == nil {
		progress = func(string, string, int) {}
	}
	node := resolveNode()
	npm := resolveNpm()
	if node == "" || npm == "" {
		return "", fmt.Errorf("Node.js/npm not found after prerequisites step")
	}

	env := os.Environ()
	// Prefer offline-friendly npm
	env = append(env, "ELECTRON_GET_USE_PROXY=0")

	progress("Dependencies", "Downloading app libraries (npm). This can take several minutes…", 72)
	logx.Log("npm ci in %s", sourceRoot)
	if err := run(sourceRoot, env, progress, npm, "ci"); err != nil {
		logx.Log("npm ci failed, trying npm install: %v", err)
		progress("Dependencies", "Retrying with npm install…", 73)
		if err2 := run(sourceRoot, env, progress, npm, "install"); err2 != nil {
			return "", fmt.Errorf("npm install failed: %w", err2)
		}
	}

	// Verify electron binary
	progress("Electron", "Checking Electron runtime binary…", 78)
	electronExe := filepath.Join(sourceRoot, "node_modules", "electron", "dist", "electron.exe")
	if !fileExists(electronExe) {
		logx.Log("electron.exe missing — running electron/install.js")
		progress("Electron", "Downloading Electron (required). Please wait…", 79)
		installJS := filepath.Join(sourceRoot, "node_modules", "electron", "install.js")
		if !fileExists(installJS) {
			return "", fmt.Errorf("Electron package missing after npm install")
		}
		if err := run(sourceRoot, env, progress, node, installJS); err != nil {
			return "", fmt.Errorf("Electron download failed: %w\n\nOften caused by network blocks. Retry later.", err)
		}
		if !fileExists(electronExe) {
			return "", fmt.Errorf("Electron binary still missing at %s", electronExe)
		}
	}
	logx.Log("electron.exe OK: %s", electronExe)
	progress("Electron", "Electron runtime ready.", 82)

	// Production build
	progress("Build", "Building Slate (compile). This may take a few minutes…", 84)
	if err := run(sourceRoot, env, progress, npm, "run", "build"); err != nil {
		return "", fmt.Errorf("npm run build failed: %w", err)
	}
	// verify out/main
	mainJS := filepath.Join(sourceRoot, "out", "main", "index.js")
	if !fileExists(mainJS) {
		return "", fmt.Errorf("build output missing: %s", mainJS)
	}
	// Confirm IS_WIN patch survived into bundle if possible
	if b, err := os.ReadFile(mainJS); err == nil {
		if strings.Contains(string(b), "IS_WIN") || strings.Contains(string(b), "winShimTarget") || strings.Contains(string(b), "WIN_CLI") {
			logx.Log("Patch marker found in out/main/index.js")
		} else {
			logx.Log("Warning: patch marker not found in bundled main (minification may rename)")
		}
	}

	// Package with electron-packager
	progress("Package", "Creating Windows app folder…", 90)
	outDir := filepath.Join(sourceRoot, "dist-win")
	_ = os.RemoveAll(outDir)

	// Prefer npx @electron/packager
	npx := resolveNpx()
	icon := findIcon(sourceRoot)
	args := []string{
		"--yes", "@electron/packager", ".", "Slate",
		"--platform=win32", "--arch=x64",
		"--out=" + outDir,
		"--overwrite",
		"--prune=true",
		"--ignore=dist-win",
		"--ignore=\\.git",
	}
	if icon != "" {
		args = append(args, "--icon="+icon)
	}
	logx.Log("packager: %v", args)
	if err := run(sourceRoot, env, progress, npx, args...); err != nil {
		// try local bin
		packager := filepath.Join(sourceRoot, "node_modules", ".bin", "electron-packager.cmd")
		if fileExists(packager) {
			if err2 := run(sourceRoot, env, progress, packager, args[2:]...); err2 != nil {
				return "", fmt.Errorf("packaging failed: %w", err)
			}
		} else {
			// install packager then retry
			progress("Package", "Installing packager helper…", 91)
			_ = run(sourceRoot, env, progress, npm, "install", "--no-save", "@electron/packager")
			if err3 := run(sourceRoot, env, progress, npx, args...); err3 != nil {
				return "", fmt.Errorf("packaging failed: %w", err3)
			}
		}
	}

	// Find Slate.exe
	var exePath string
	_ = filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.EqualFold(info.Name(), "Slate.exe") {
			exePath = path
			return filepath.SkipAll
		}
		return nil
	})
	if exePath == "" {
		return "", fmt.Errorf("packaged Slate.exe not found under %s", outDir)
	}
	appDir := filepath.Dir(exePath)
	logx.Log("Packaged app dir: %s", appDir)
	progress("Package", "Windows app package ready.", 93)
	return appDir, nil
}

func run(dir string, env []string, progress ProgressFn, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = env
	logx.Log("RUN %s %s", name, strings.Join(args, " "))
	out, err := cmd.CombinedOutput()
	logx.Log("OUT (%s): %s", name, truncate(string(out), 4000))
	if err != nil {
		return fmt.Errorf("%v: %s", err, truncate(string(out), 800))
	}
	return nil
}

func resolveNode() string {
	if p, err := exec.LookPath("node"); err == nil {
		return p
	}
	c := `C:\Program Files\nodejs\node.exe`
	if fileExists(c) {
		return c
	}
	return ""
}

func resolveNpm() string {
	if p, err := exec.LookPath("npm.cmd"); err == nil {
		return p
	}
	if p, err := exec.LookPath("npm"); err == nil {
		return p
	}
	c := `C:\Program Files\nodejs\npm.cmd`
	if fileExists(c) {
		return c
	}
	return ""
}

func resolveNpx() string {
	if p, err := exec.LookPath("npx.cmd"); err == nil {
		return p
	}
	if p, err := exec.LookPath("npx"); err == nil {
		return p
	}
	c := `C:\Program Files\nodejs\npx.cmd`
	if fileExists(c) {
		return c
	}
	return "npx"
}

func findIcon(root string) string {
	cands := []string{
		filepath.Join(root, "build", "icon.ico"),
		filepath.Join(root, "resources", "icon.ico"),
		filepath.Join(root, "build", "icon.png"),
	}
	for _, c := range cands {
		if fileExists(c) {
			return c
		}
	}
	return ""
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// Ensure node path is used even if unused import
var _ = time.Second
