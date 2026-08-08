package source

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wassermanproductions/slate-installer/internal/logx"
	"github.com/wassermanproductions/slate-installer/internal/paths"
)

type ProgressFn func(step, detail string, percent int)

// Ensure downloads and extracts Slate source if missing or force.
func Ensure(progress ProgressFn, force bool) (string, error) {
	if progress == nil {
		progress = func(string, string, int) {}
	}
	src := paths.SourceDir()
	marker := filepath.Join(src, "package.json")
	if !force && fileExists(marker) {
		logx.Log("Source already present: %s", src)
		progress("Source", "Slate source already downloaded.", 65)
		return src, nil
	}

	if err := os.MkdirAll(paths.WorkDir(), 0o755); err != nil {
		return "", err
	}
	zipPath := filepath.Join(paths.WorkDir(), "slate-"+paths.SourceRef+".zip")

	progress("Source", "Downloading Slate from GitHub…", 62)
	logx.Log("Fetching %s", paths.GitHubZipURL)
	if err := download(paths.GitHubZipURL, zipPath, 10*time.Minute); err != nil {
		// Fallback: local slate-0.3.2 next to installer workspace
		local := findLocalSnapshot()
		if local != "" {
			logx.Log("Network download failed (%v); using local snapshot %s", err, local)
			progress("Source", "Using local Slate source on this PC…", 64)
			if err2 := copyDir(local, src); err2 != nil {
				return "", fmt.Errorf("download failed (%v) and local copy failed (%v)", err, err2)
			}
			return src, nil
		}
		return "", fmt.Errorf("could not download Slate source: %w\n\nCheck your internet connection and try again.", err)
	}

	progress("Source", "Unpacking…", 68)
	// clean target
	_ = os.RemoveAll(src)
	if err := unzip(zipPath, paths.WorkDir()); err != nil {
		return "", fmt.Errorf("extract zip: %w", err)
	}
	// GitHub zips as slate-<branch> (e.g. slate-windows-support)
	if !fileExists(marker) {
		entries, _ := os.ReadDir(paths.WorkDir())
		for _, e := range entries {
			if e.IsDir() && strings.HasPrefix(e.Name(), "slate-") {
				cand := filepath.Join(paths.WorkDir(), e.Name())
				if fileExists(filepath.Join(cand, "package.json")) {
					src = cand
					break
				}
			}
		}
	}
	if !fileExists(filepath.Join(src, "package.json")) {
		return "", fmt.Errorf("source package.json missing after extract")
	}
	logx.Log("Source ready at %s", src)
	progress("Source", "Source ready.", 70)
	return src, nil
}

func findLocalSnapshot() string {
	// installer lives in .../slate/slate-installer; sibling slate-0.3.2
	candidates := []string{}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(dir, "..", "slate-0.3.2"),
			filepath.Join(dir, "..", "..", "slate-0.3.2"),
			filepath.Join(dir, "slate-0.3.2"),
		)
	}
	wd, _ := os.Getwd()
	candidates = append(candidates,
		filepath.Join(wd, "..", "slate-0.3.2"),
		filepath.Join(wd, "slate-0.3.2"),
		`M:\Users\Chris\Documents\_code-projects\slate\slate-0.3.2`,
	)
	for _, c := range candidates {
		c = filepath.Clean(c)
		if fileExists(filepath.Join(c, "package.json")) {
			return c
		}
	}
	return ""
}

func download(url, dest string, timeout time.Duration) error {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "SlateSetup/1.0")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("HTTP %d from GitHub", res.StatusCode)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, res.Body)
	return err
}

func unzip(zipPath, dest string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		// prevent zip slip
		target := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)+string(os.PathSeparator)) &&
			filepath.Clean(target) != filepath.Clean(dest) {
			return fmt.Errorf("illegal path in zip: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(target, 0o755)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		// skip node_modules if present
		if strings.Contains(rel, "node_modules") || strings.Contains(rel, ".git") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		_ = os.MkdirAll(filepath.Dir(target), 0o755)
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
