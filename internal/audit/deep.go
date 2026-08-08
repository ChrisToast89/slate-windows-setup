package audit

import (
	"github.com/wassermanproductions/slate-installer/internal/installstatus"
	"github.com/wassermanproductions/slate-installer/internal/logx"
	"github.com/wassermanproductions/slate-installer/internal/update"
)

// DeepReport is the full audit-tool view: PC + install health + update check + projects.
type DeepReport struct {
	System     Report                 `json:"system"`
	Install    installstatus.Status   `json:"install"`
	Updates    update.CheckResult     `json:"updates"`
	Summary    string                 `json:"summary"`
	Highlights []string               `json:"highlights"`
}

// RunDeep is the stand-alone Audit tool: no system changes, read-only.
func RunDeep() DeepReport {
	logx.Log("Deep audit (read-only)")
	sys := Run()
	inst := installstatus.Detect()
	upd := update.Check()

	highlights := []string{}
	if inst.Installed && inst.InstallHealthy {
		highlights = append(highlights, "✓ Slate is installed and looks healthy")
	} else if inst.Installed {
		highlights = append(highlights, "⚠ Slate files found but install may be incomplete — use Repair")
	} else {
		highlights = append(highlights, "○ Slate is not installed yet")
	}
	highlights = append(highlights,
		"Your projects: "+inst.ProjectsDir+
			ternary(inst.ProjectsDirExists, "", " (not created yet)"),
	)
	if inst.ProjectsDirExists {
		highlights = append(highlights, ternary(inst.ProjectCount > 0,
			"Project folders: counted read-only (never modified by Setup)",
			"Projects folder exists but no projects yet"))
	}
	highlights = append(highlights, "Protection policy: Setup never deletes "+inst.ProjectsDir)

	if upd.Online {
		if upd.UpdateAvailable {
			highlights = append(highlights, "↑ Update available from GitHub — use Install update (projects stay safe)")
		} else if upd.UpToDate {
			highlights = append(highlights, "✓ App matches latest repository source")
		} else {
			highlights = append(highlights, upd.Summary)
		}
	} else if upd.Error != "" {
		highlights = append(highlights, "Could not reach GitHub for updates: "+upd.Error)
	}

	if !sys.NodeOK {
		highlights = append(highlights, "Node.js 20+ missing (needed to build/update Slate)")
	}
	if !sys.FFmpegOK {
		highlights = append(highlights, "ffmpeg not on PATH (optional; media features)")
	}

	summary := inst.Summary
	if upd.UpdateAvailable {
		summary += " " + upd.Summary
	}

	return DeepReport{
		System:     sys,
		Install:    inst,
		Updates:    upd,
		Summary:    summary,
		Highlights: highlights,
	}
}
