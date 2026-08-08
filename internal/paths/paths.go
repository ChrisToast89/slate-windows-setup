package paths

import (
	"os"
	"path/filepath"
	"strings"
)

// Pin source to the community Windows-support fork until patches land upstream.
// App authorship remains Sam Wasserman (wassermanproductions/slate).
// This Setup tool downloads/builds from ChrisToast89/slate@windows-support.
const SourceRef = "windows-support"

const (
	// Fork used for Windows-ready source (includes brain.ts CLI fix).
	GitHubOwner = "ChrisToast89"
	GitHubRepo  = "slate"
	// Official app upstream (credits, releases display).
	UpstreamOwner = "wassermanproductions"
	UpstreamRepo  = "slate"

	GitHubZipURL = "https://github.com/" + GitHubOwner + "/" + GitHubRepo + "/archive/refs/heads/" + SourceRef + ".zip"
	// Commits API for update checks (windows-support tip on the fork).
	GitHubCommitsAPI  = "https://api.github.com/repos/" + GitHubOwner + "/" + GitHubRepo + "/commits/" + SourceRef
	GitHubReleasesAPI = "https://api.github.com/repos/" + UpstreamOwner + "/" + UpstreamRepo + "/releases/latest"
)

const InstallerVersion = "1.1.2"

func LocalAppData() string {
	if d := os.Getenv("LOCALAPPDATA"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "AppData", "Local")
}

func TempLog() string {
	return filepath.Join(os.TempDir(), "slate-install.log")
}

// WorkDir holds downloaded source + build artifacts (never user projects).
func WorkDir() string {
	return filepath.Join(LocalAppData(), "Slate", "build-work")
}

// SourceDir is where the extracted GitHub zip lands (slate-main).
func SourceDir() string {
	return filepath.Join(WorkDir(), "slate-"+SourceRef)
}

// InstallDir is the per-user app install location (app binaries only).
func InstallDir() string {
	return filepath.Join(LocalAppData(), "Programs", "Slate")
}

func InstalledExe() string {
	return filepath.Join(InstallDir(), "Slate.exe")
}

// ManifestPath records what was installed (for update checks). Lives under install dir only.
func ManifestPath() string {
	return filepath.Join(InstallDir(), "slate-install-manifest.json")
}

// ProjectsDir is where the Slate app stores user project JSON.
// MUST NEVER be deleted, overwritten, or modified by the installer/updater.
func ProjectsDir() string {
	if d := os.Getenv("SLATE_DATA_DIR"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Documents", "Slate")
}

func StartMenuShortcut() string {
	programs := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs")
	return filepath.Join(programs, "Slate.lnk")
}

func DesktopShortcut() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Desktop", "Slate.lnk")
}

// IsProtectedPath reports whether p is under a path the installer must never wipe.
func IsProtectedPath(p string) bool {
	p = filepath.Clean(p)
	protected := []string{
		ProjectsDir(),
		filepath.Join(os.Getenv("USERPROFILE"), "Documents", "Slate"),
	}
	for _, root := range protected {
		root = filepath.Clean(root)
		if root == "" || root == "." {
			continue
		}
		if strings.EqualFold(p, root) {
			return true
		}
		rel, err := filepath.Rel(root, p)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

// AssertSafeToReplaceInstallDir refuses if install dir somehow collides with projects.
func AssertSafeToReplaceInstallDir() error {
	inst := filepath.Clean(InstallDir())
	proj := filepath.Clean(ProjectsDir())
	if strings.EqualFold(inst, proj) {
		return errProtected("install directory equals projects directory")
	}
	if IsProtectedPath(inst) {
		return errProtected("install directory is under protected projects path")
	}
	// Refuse if projects is inside install (misconfiguration)
	if rel, err := filepath.Rel(inst, proj); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return errProtected("projects appear to live under install directory — aborting to protect data")
	}
	return nil
}

type protectError string

func (e protectError) Error() string { return string(e) }

func errProtected(msg string) error {
	return protectError("SAFETY STOP: " + msg + ". User project files were not touched.")
}
