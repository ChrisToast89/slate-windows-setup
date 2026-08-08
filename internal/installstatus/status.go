package installstatus

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/wassermanproductions/slate-installer/internal/logx"
	"github.com/wassermanproductions/slate-installer/internal/paths"
)

// Flavor of a discovered Slate install.
const (
	FlavorNone     = ""
	FlavorPackaged = "packaged" // %LOCALAPPDATA%\Programs\Slate (Setup install)
	FlavorNPM      = "npm"      // electron-vite source tree (npm run dev / build + electron)
)

// Manifest is written after a successful Setup install/update (app folder only).
type Manifest struct {
	InstallerVersion  string `json:"installerVersion"`
	SourceRef         string `json:"sourceRef"`
	SourceSHA         string `json:"sourceSha"`
	ReleaseTag        string `json:"releaseTag,omitempty"`
	AppVersion        string `json:"appVersion,omitempty"`
	InstalledAt       string `json:"installedAt"`
	UpdatedAt         string `json:"updatedAt"`
	ExePath           string `json:"exePath"`
	InstallDir        string `json:"installDir"`
	SmokeOK           bool   `json:"smokeOk"`
	ProjectsPath      string `json:"projectsPath"`
	ProjectsPreserved bool   `json:"projectsPreserved"`
	Flavor            string `json:"flavor,omitempty"`
}

// Instance is one discovered working copy of Slate.
type Instance struct {
	Flavor      string `json:"flavor"` // packaged | npm
	Root        string `json:"root"`
	ExePath     string `json:"exePath,omitempty"`
	LaunchHint  string `json:"launchHint"` // how a human runs it
	Version     string `json:"version,omitempty"`
	Healthy     bool   `json:"healthy"`
	Description string `json:"description"`
}

// Status is a human-friendly snapshot of installs on this machine.
type Status struct {
	Installed         bool        `json:"installed"`
	InstallHealthy    bool        `json:"installHealthy"`
	Flavor            string      `json:"flavor"` // primary instance flavor
	ExePath           string      `json:"exePath"`
	InstallDir        string      `json:"installDir"`
	ExeExists         bool        `json:"exeExists"`
	ManifestExists    bool        `json:"manifestExists"`
	Manifest          *Manifest   `json:"manifest,omitempty"`
	ShortcutOK        bool        `json:"shortcutOk"`
	UninstallRegOK    bool        `json:"uninstallRegOk"`
	ProjectsDir       string      `json:"projectsDir"`
	ProjectsDirExists bool        `json:"projectsDirExists"`
	ProjectCount      int         `json:"projectCount"`
	ProjectsProtected bool        `json:"projectsProtected"`
	Instances         []Instance  `json:"instances"`
	Summary           string      `json:"summary"`
	DetailLines       []string    `json:"detailLines"`
}

// Detect whether Slate is installed / runnable on this machine.
// Recognizes:
//  1. Packaged Setup install → %LOCALAPPDATA%\Programs\Slate
//  2. NPM/Electron source trees (package.json name "slate" + electron + out/main)
//
// Explicitly ignores slate-windows (Wails) trees.
func Detect() Status {
	var s Status
	s.InstallDir = paths.InstallDir()
	s.ExePath = paths.InstalledExe()
	s.ProjectsDir = paths.ProjectsDir()
	s.ProjectsProtected = true
	s.ShortcutOK = fileExists(paths.StartMenuShortcut())
	s.UninstallRegOK = uninstallKeyExists()
	s.ProjectsDirExists = dirExists(s.ProjectsDir)
	s.ProjectCount = countProjects(s.ProjectsDir)

	seen := map[string]bool{}
	add := func(inst Instance) {
		key := strings.ToLower(filepath.Clean(inst.Root))
		if key == "" || seen[key] {
			return
		}
		seen[key] = true
		s.Instances = append(s.Instances, inst)
	}

	// --- 1) Packaged install (what Setup writes) ---
	if m, err := ReadManifest(); err == nil && m != nil {
		s.ManifestExists = true
		s.Manifest = m
	}
	packagedExe := paths.InstalledExe()
	if fileExists(packagedExe) {
		add(Instance{
			Flavor:      FlavorPackaged,
			Root:        paths.InstallDir(),
			ExePath:     packagedExe,
			LaunchHint:  packagedExe,
			Version:     packageVersion(paths.InstallDir()),
			Healthy:     true,
			Description: "Installed by Slate Setup (Start Menu / Programs folder)",
		})
	} else if dirExists(paths.InstallDir()) {
		if found := findExe(paths.InstallDir()); found != "" {
			add(Instance{
				Flavor:      FlavorPackaged,
				Root:        filepath.Dir(found),
				ExePath:     found,
				LaunchHint:  found,
				Healthy:     true,
				Description: "Installed by Slate Setup (packaged app folder)",
			})
		}
	}
	if s.Manifest != nil && s.Manifest.ExePath != "" && fileExists(s.Manifest.ExePath) {
		add(Instance{
			Flavor:      FlavorPackaged,
			Root:        filepath.Dir(s.Manifest.ExePath),
			ExePath:     s.Manifest.ExePath,
			LaunchHint:  s.Manifest.ExePath,
			Healthy:     s.Manifest.SmokeOK || true,
			Description: "Recorded in Setup install manifest",
		})
	}

	// --- 2) NPM / electron-vite working trees ---
	for _, root := range npmCandidateRoots() {
		if inst, ok := probeNPMTree(root); ok {
			add(inst)
		}
	}

	// Primary = prefer packaged, else first healthy npm
	var primary *Instance
	for i := range s.Instances {
		if s.Instances[i].Flavor == FlavorPackaged && s.Instances[i].Healthy {
			primary = &s.Instances[i]
			break
		}
	}
	if primary == nil {
		for i := range s.Instances {
			if s.Instances[i].Healthy {
				primary = &s.Instances[i]
				break
			}
		}
	}

	if primary != nil {
		s.Installed = true
		s.InstallHealthy = primary.Healthy
		s.Flavor = primary.Flavor
		s.ExePath = primary.ExePath
		s.InstallDir = primary.Root
		s.ExeExists = primary.ExePath != "" && fileExists(primary.ExePath)
	}

	// Summary
	lines := []string{}
	if !s.Installed {
		s.Summary = "Slate is not installed yet on this PC."
		lines = append(lines,
			"Setup looks for:",
			"  • Packaged app: "+paths.InstallDir(),
			"  • Or an npm/Electron tree (folder with package.json name \"slate\", node_modules/electron, and out/main)",
		)
	} else if s.Flavor == FlavorNPM {
		s.Summary = "Found a working Slate (npm / Electron) install."
		lines = append(lines,
			"Type: Development / npm (electron-vite)",
			"Folder: "+s.InstallDir,
			"Launch: from that folder run  npm run dev  or  npm start",
			"Electron: "+filepath.Join(s.InstallDir, "node_modules", "electron", "dist", "electron.exe"),
		)
		if s.ExePath != "" {
			lines = append(lines, "Electron binary: present")
		}
		// Note if Setup packaged install is missing
		if !fileExists(paths.InstalledExe()) {
			lines = append(lines,
				"",
				"Note: This is your source/npm copy — not the Programs\\Slate package yet.",
				"Slate Setup can still install a Start Menu copy; your projects stay in Documents\\Slate either way.",
			)
		}
	} else {
		s.Summary = "Slate is installed and looks healthy."
		lines = append(lines, "Type: Packaged (Slate Setup)", "App: "+s.ExePath)
		if s.Manifest != nil && s.Manifest.SourceSHA != "" {
			lines = append(lines, "Installed source: "+shortSHA(s.Manifest.SourceSHA)+" ("+s.Manifest.SourceRef+")")
		}
	}

	if len(s.Instances) > 1 {
		lines = append(lines, "", fmt.Sprintf("Found %d Slate locations:", len(s.Instances)))
		for _, inst := range s.Instances {
			lines = append(lines, fmt.Sprintf("  • [%s] %s", inst.Flavor, inst.Root))
		}
	}

	lines = append(lines, "", "Your projects folder: "+s.ProjectsDir)
	if s.ProjectsDirExists {
		lines = append(lines, fmt.Sprintf("Project folders found: %d (read-only — never modified by Setup)", s.ProjectCount))
	} else {
		lines = append(lines, "No projects folder yet (created when you use Slate)")
	}
	lines = append(lines, "Protection: Setup never deletes or overwrites "+s.ProjectsDir)
	s.DetailLines = lines
	logx.Log("Install status: installed=%v healthy=%v flavor=%s instances=%d", s.Installed, s.InstallHealthy, s.Flavor, len(s.Instances))
	return s
}

// probeNPMTree returns true if root is a runnable Electron Slate source tree.
func probeNPMTree(root string) (Instance, bool) {
	root = filepath.Clean(root)
	if root == "" || isIgnoredTree(root) {
		return Instance{}, false
	}
	pkgPath := filepath.Join(root, "package.json")
	if !fileExists(pkgPath) {
		return Instance{}, false
	}
	name, ver := readPackageNameVersion(pkgPath)
	if !strings.EqualFold(name, "slate") {
		return Instance{}, false
	}
	// Must look like Electron slate, not a random package named slate
	mainJS := filepath.Join(root, "out", "main", "index.js")
	electronExe := filepath.Join(root, "node_modules", "electron", "dist", "electron.exe")
	hasSrcMain := fileExists(filepath.Join(root, "src", "main", "index.ts")) || fileExists(filepath.Join(root, "src", "main", "index.js"))
	hasElectronVite := fileExists(filepath.Join(root, "electron.vite.config.ts")) || fileExists(filepath.Join(root, "electron.vite.config.js"))

	if !hasSrcMain && !hasElectronVite && !fileExists(mainJS) {
		return Instance{}, false
	}

	healthy := fileExists(electronExe) && (fileExists(mainJS) || hasSrcMain)
	// Built production is stronger signal
	if fileExists(mainJS) && fileExists(electronExe) {
		healthy = true
	}

	desc := "npm / Electron source tree"
	if fileExists(mainJS) {
		desc += " (built: out/main present)"
	} else {
		desc += " (source present; run npm run build if needed)"
	}
	if !fileExists(electronExe) {
		desc += " — electron binary missing (run npm install)"
		healthy = false
	}

	return Instance{
		Flavor:      FlavorNPM,
		Root:        root,
		ExePath:     ternary(fileExists(electronExe), electronExe, ""),
		LaunchHint:  "cd \"" + root + "\" && npm run dev",
		Version:     ver,
		Healthy:     healthy,
		Description: desc,
	}, healthy || fileExists(mainJS) || hasSrcMain
}

// isIgnoredTree skips Wails port and installer itself.
func isIgnoredTree(root string) bool {
	base := strings.ToLower(filepath.Base(root))
	if base == "slate-windows" || base == "slate-installer" {
		return true
	}
	// Wails signature
	if fileExists(filepath.Join(root, "wails.json")) && fileExists(filepath.Join(root, "go.mod")) {
		return true
	}
	if fileExists(filepath.Join(root, "frontend", "wailsjs", "go", "main", "App.js")) {
		return true
	}
	return false
}

func npmCandidateRoots() []string {
	var out []string
	add := func(p string) {
		p = filepath.Clean(p)
		if p != "" && p != "." {
			out = append(out, p)
		}
	}

	// Installer's own download/build workdir
	add(paths.SourceDir())
	add(filepath.Join(paths.WorkDir(), "slate-main"))
	add(filepath.Join(paths.LocalAppData(), "Slate", "app"))

	// Relative to this Setup executable
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		// .../slate-installer/build/bin → ../../../slate
		add(filepath.Join(dir, "..", "..", "..", "slate"))
		add(filepath.Join(dir, "..", "..", "..", "slate-0.3.2"))
		add(filepath.Join(dir, "..", "..", "slate"))
		add(filepath.Join(dir, "..", "slate"))
	}

	// Relative to cwd
	if wd, err := os.Getwd(); err == nil {
		add(filepath.Join(wd, "slate"))
		add(filepath.Join(wd, "..", "slate"))
		add(filepath.Join(wd, "..", "slate-0.3.2"))
		add(wd) // if user runs Setup from the slate source folder
	}

	// Common user code locations
	home, _ := os.UserHomeDir()
	docs := filepath.Join(home, "Documents")
	for _, base := range []string{
		filepath.Join(docs, "_code-projects", "slate"),
		filepath.Join(docs, "code-projects", "slate"),
		filepath.Join(docs, "GitHub", "slate"),
		filepath.Join(docs, "github", "slate"),
		filepath.Join(docs, "projects", "slate"),
		filepath.Join(home, "source", "slate"),
		filepath.Join(home, "src", "slate"),
		filepath.Join(home, "dev", "slate"),
	} {
		add(filepath.Join(base, "slate"))
		add(filepath.Join(base, "slate-0.3.2"))
		add(base) // clone named "slate" at root
	}

	// Drive letter scan for this workspace pattern (cheap fixed path)
	add(`M:\Users\Chris\Documents\_code-projects\slate\slate`)
	add(`M:\Users\Chris\Documents\_code-projects\slate\slate-0.3.2`)

	return out
}

func readPackageNameVersion(pkgPath string) (name, version string) {
	raw, err := os.ReadFile(pkgPath)
	if err != nil {
		return "", ""
	}
	var meta struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if json.Unmarshal(raw, &meta) != nil {
		return "", ""
	}
	return meta.Name, meta.Version
}

func packageVersion(root string) string {
	_, v := readPackageNameVersion(filepath.Join(root, "package.json"))
	return v
}

func ReadManifest() (*Manifest, error) {
	raw, err := os.ReadFile(paths.ManifestPath())
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// WriteManifest stores install metadata next to the packaged app (not near projects).
func WriteManifest(m Manifest) error {
	if err := paths.AssertSafeToReplaceInstallDir(); err != nil {
		return err
	}
	m.ProjectsPath = paths.ProjectsDir()
	m.ProjectsPreserved = true
	m.Flavor = FlavorPackaged
	if m.InstalledAt == "" {
		m.InstalledAt = time.Now().UTC().Format(time.RFC3339)
	}
	m.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	m.InstallerVersion = paths.InstallerVersion
	m.InstallDir = paths.InstallDir()
	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(paths.ManifestPath(), raw, 0o644)
}

func countProjects(root string) int {
	entries, err := os.ReadDir(root)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if fileExists(filepath.Join(root, e.Name(), "project.json")) {
			n++
		}
	}
	return n
}

func findExe(root string) string {
	var found string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.EqualFold(info.Name(), "Slate.exe") {
			// skip wails build bins under slate-windows
			if strings.Contains(strings.ToLower(path), "slate-windows") {
				return nil
			}
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func uninstallKeyExists() bool {
	cmd := exec.Command("reg", "query", `HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\Slate`)
	return cmd.Run() == nil
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

func dirExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

func shortSHA(s string) string {
	if len(s) > 7 {
		return s[:7]
	}
	return s
}

func ternary(c bool, a, b string) string {
	if c {
		return a
	}
	return b
}
