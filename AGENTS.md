# AGENTS — slate-installer (Slate Setup for Windows)

## What this repository is

**Slate Setup for Windows** — a **helper installer** so users can run **Sam Wasserman’s original Electron/npm Slate** on Windows without hand-running Node/npm builds.

- Does **not** rebrand Slate as a new app  
- Does **not** replace Slate with a Wails port  
- Downloads/builds **upstream Electron Slate**, installs to the user profile  
- Public packaging name: **slate-windows-setup**  
- GitHub: https://github.com/ChrisToast89/slate-windows-setup  
- Typical install: `%LOCALAPPDATA%\Programs\Slate`

Upstream: https://github.com/wassermanproductions/slate (Apache-2.0, credit Sam Wasserman).

## What this is NOT

| Not this | That lives at |
|----------|----------------|
| **Win-Slate** (ported Go+Wails app) | Workspace `Slate-win/Win-Slate/` · GitHub `ChrisToast89/Win-Slate` |
| Legacy early Wails port tree | Workspace `Slate-win/slate-windows/` |
| A reimplementation of the prompt studio UI | That is Win-Slate / upstream |

## Naming

| Name | Meaning |
|------|---------|
| Folder **`slate-installer`** | This source tree (local) |
| Product **Slate Setup for Windows** / **SlateSetup.exe** | What users run |
| GitHub **slate-windows-setup** | Release repo name |
| Phrase **“slate-windows”** (installer sense) | **This product**, not the Wails port folder |

## Layout (short)

- `app.go` / `main.go` / `internal/` — Setup logic (audit, prereqs, source fetch, build, install)  
- `frontend/` — Setup UI  
- `dist/` — packaged zips for releases  
- `docs/USAGE.md` — operator notes  

## When to edit here

- Install flow for **Electron Slate** on Windows  
- Node/ffmpeg/Claude Code helpers for that install  
- SmartScreen packaging for **SlateSetup**  

## When to leave this tree alone

- Bugs in the **ported** Wails UI/brain → `Win-Slate`  
- Changing Slate’s creative features → upstream Electron source or Win-Slate, not Setup  

## Coexistence with Win-Slate

- Electron install: **Programs\Slate**  
- Port install: **Programs\Win-Slate**  
- Projects: often shared under **Documents\Slate**  
- Do not uninstall or overwrite the other product when improving Setup  

## Workspace map

See `../PRODUCT-MAP.md` (workspace root).
