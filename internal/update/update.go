package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/wassermanproductions/slate-installer/internal/installstatus"
	"github.com/wassermanproductions/slate-installer/internal/logx"
	"github.com/wassermanproductions/slate-installer/internal/paths"
)

// CheckResult compares the installed app to the GitHub repository.
type CheckResult struct {
	CheckedAt        string `json:"checkedAt"`
	Online           bool   `json:"online"`
	Error            string `json:"error,omitempty"`
	LatestSHA        string `json:"latestSha"`
	LatestMessage    string `json:"latestMessage"`
	LatestDate       string `json:"latestDate"`
	LatestReleaseTag string `json:"latestReleaseTag,omitempty"`
	InstalledSHA     string `json:"installedSha"`
	Installed        bool   `json:"installed"`
	UpdateAvailable  bool   `json:"updateAvailable"`
	UpToDate         bool   `json:"upToDate"`
	Summary          string `json:"summary"`
	Detail           string `json:"detail"`
	// Safety note always present for UI
	ProjectsNote string `json:"projectsNote"`
}

// Check queries GitHub for the latest main commit (and latest release tag for display).
func Check() CheckResult {
	r := CheckResult{
		CheckedAt:    time.Now().UTC().Format(time.RFC3339),
		ProjectsNote: "Updates only replace the app program files. Your projects in " + paths.ProjectsDir() + " are never deleted or overwritten.",
	}

	st := installstatus.Detect()
	r.Installed = st.Installed
	if st.Manifest != nil {
		r.InstalledSHA = st.Manifest.SourceSHA
	}

	client := &http.Client{Timeout: 25 * time.Second}

	// Latest commit on tracked ref
	sha, msg, date, err := fetchCommit(client, paths.GitHubCommitsAPI)
	if err != nil {
		r.Online = false
		r.Error = err.Error()
		r.Summary = "Could not check for updates (network or GitHub unavailable)."
		r.Detail = err.Error()
		logx.Log("update check failed: %v", err)
		return r
	}
	r.Online = true
	r.LatestSHA = sha
	r.LatestMessage = msg
	r.LatestDate = date

	// Latest release tag (informational)
	if tag, err := fetchLatestTag(client, paths.GitHubReleasesAPI); err == nil {
		r.LatestReleaseTag = tag
	}

	if !r.Installed {
		r.UpdateAvailable = false
		r.UpToDate = false
		r.Summary = "Slate is not installed. Use Install to set it up."
		r.Detail = fmt.Sprintf("Latest on GitHub %s: %s — %s", paths.SourceRef, short(sha), msg)
		return r
	}

	if r.InstalledSHA == "" {
		// Installed but no manifest (older install) — treat as update recommended
		r.UpdateAvailable = true
		r.UpToDate = false
		r.Summary = "An update is recommended (this install has no version record yet)."
		r.Detail = fmt.Sprintf("Latest source %s. Updating reinstalls the app only; your projects stay safe.", short(sha))
		return r
	}

	if strings.EqualFold(r.InstalledSHA, r.LatestSHA) {
		r.UpdateAvailable = false
		r.UpToDate = true
		r.Summary = "You already have the latest Slate from the repository."
		r.Detail = fmt.Sprintf("Installed and latest: %s", short(sha))
		return r
	}

	r.UpdateAvailable = true
	r.UpToDate = false
	r.Summary = "A newer version is available on GitHub."
	r.Detail = fmt.Sprintf("Installed: %s → Latest: %s (%s)", short(r.InstalledSHA), short(r.LatestSHA), msg)
	logx.Log("update available: %s -> %s", r.InstalledSHA, r.LatestSHA)
	return r
}

func fetchCommit(client *http.Client, url string) (sha, message, date string, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("User-Agent", "SlateSetup/"+paths.InstallerVersion)
	req.Header.Set("Accept", "application/vnd.github+json")
	res, err := client.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("network error contacting GitHub: %w", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		return "", "", "", fmt.Errorf("GitHub returned HTTP %d", res.StatusCode)
	}
	var parsed struct {
		SHA    string `json:"sha"`
		Commit struct {
			Message string `json:"message"`
			Author  struct {
				Date string `json:"date"`
			} `json:"author"`
		} `json:"commit"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", "", "", err
	}
	if parsed.SHA == "" {
		return "", "", "", fmt.Errorf("empty commit from GitHub")
	}
	msg := strings.TrimSpace(parsed.Commit.Message)
	if i := strings.Index(msg, "\n"); i > 0 {
		msg = msg[:i]
	}
	return parsed.SHA, msg, parsed.Commit.Author.Date, nil
}

func fetchLatestTag(client *http.Client, url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "SlateSetup/"+paths.InstallerVersion)
	req.Header.Set("Accept", "application/vnd.github+json")
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", res.StatusCode)
	}
	var parsed struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return "", err
	}
	return parsed.TagName, nil
}

func short(s string) string {
	if len(s) > 7 {
		return s[:7]
	}
	return s
}
