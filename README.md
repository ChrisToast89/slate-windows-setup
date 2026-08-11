# Slate Setup for Windows

> **Product identity (maintainers / agents)**  
> This tree is the **installer for Sam’s original Electron/npm Slate** only.  
> Local folder: `slate-installer`. Published as **slate-windows-setup**.  
> It is **not** Win-Slate (the Go+Wails port) and **not** the archived early port under `_archive/`.  
> See [AGENTS.md](./AGENTS.md) and workspace [PRODUCT-MAP.md](../PRODUCT-MAP.md).

A **helper installer** so Windows users can run **[Slate](https://github.com/wassermanproductions/slate)** without wrestling with Node, npm, and build steps.

<img width="400" alt="image" src="https://github.com/user-attachments/assets/0958ba6f-b6e5-4bb2-a81f-3f0d0f15f087" />

## Download the installer

**[SlateSetup for Windows v1.2.0 (zip)](https://github.com/ChrisToast89/slate-windows-setup/releases/download/v1.2.0/SlateSetup-windows-v1.2.0.zip)** 

All releases: [github.com/ChrisToast89/slate-windows-setup/releases](https://github.com/ChrisToast89/slate-windows-setup/releases)

### What this package is for

Slate is a **macOS-first** desktop app (Electron). There is no official Windows build from the author. This Setup helper exists so a Windows user can still get **Sam Wasserman’s** Slate onto their PC: it checks the machine, installs missing tools (e.g. Node.js, ffmpeg) when possible, downloads the official source from GitHub, applies a small Windows-only fix for the AI “brain” CLIs, builds the app, and installs it for the current user.

It does **not** replace Slate, rebrand it, or store API keys. It only automates install steps a non-developer would otherwise have to do by hand. **Internet is required** while Setup runs.

### Basic operation

1. Download and unzip the package above.
2. Double-click **`SlateSetup.exe`**. If Windows SmartScreen appears, choose **More info** → **Run anyway**.
3. On the home screen, choose **Install Slate** (or use **Audit tool** / **Check for updates** later).
4. Let it **Check this PC**, then **Continue** → **Install Slate**. First run often takes **10–20 minutes**.
5. When finished, open **Slate** from the Start Menu.

- App files: `%LOCALAPPDATA%\Programs\Slate`
- Your projects (never deleted by Setup): `%USERPROFILE%\Documents\Slate`

### AI brain setup (optional last step)

Slate does **not** use API keys. The AI “brain” runs on **your** Claude Code (or Codex) sign-in, or on a **local** model server. Setup never automates browser login for you.

After Slate is installed, the installer’s finish screen shows the same guidance. You can also do it anytime:

**Option A — Claude Code (recommended)**

1. Install [Claude Code](https://claude.com/claude-code) if you do not have it yet:
   - In Setup (finish screen): **Install Claude Code (optional)**, or
   - In a terminal: `npm install -g @anthropic-ai/claude-code`
2. Sign in (opens your browser; you approve):
   ```text
   claude auth login
   ```
3. Confirm the CLI is available: `claude --version`
4. Open **Slate** → set **Brain** to **Claude Code** (project settings / Project Bible defaults).
5. Click the **Brain** pill in the title bar to run a live connectivity test (expects a short **READY**-style reply).

If Claude Code is already installed but not signed in, only steps 2–5 are needed.

**Option B — Codex CLI**

1. Install and sign in to the Codex CLI per OpenAI/Codex docs (`codex login` or your installed CLI’s login flow).
2. In Slate, set **Brain** to **Codex**.
3. Use the title-bar Brain pill to test.

**Option C — Local model (offline)**

1. Run any OpenAI-compatible local server (Ollama, LM Studio, vLLM, llama.cpp, etc.) with a model loaded.
2. In Slate, set **Brain** to **Local model** (and endpoint/model if needed; common ports are auto-detected).
3. Test with the Brain pill. Nothing is sent off your machine for the local backend.

**Notes**

- Setup may offer **Install Claude Code** after install; it never stores credentials or runs `claude auth login` for you.
- You can use Slate without a brain for non-AI features; agent-style tools need one of the options above.

## Who made what

| | |
|--|--|
| **Slate** (the app) | Created by **[Sam Wasserman](https://wassermanproductions.com)** — [wassermanproductions/slate](https://github.com/wassermanproductions/slate), Apache-2.0 |
| **This repo** | A Windows **setup helper** only. It downloads and builds *his* software for Windows users. It is **not** a fork of Slate and **not** an official Wasserman release. |

Credit and support for the app belong with Sam. This tool only makes Windows install less painful.

## What it does

1. Checks the PC (Windows 10/11, disk space, tools)
2. Installs missing prerequisites when possible (Node.js, ffmpeg)
3. Downloads **official** Slate source from `wassermanproductions/slate` (`main`)
4. Applies a small **Windows-only** fix so Claude Code / Codex CLI detection works on Windows (npm `.cmd` shims)
5. Builds the Electron app and installs it under `%LOCALAPPDATA%\Programs\Slate`
6. Start Menu shortcut + uninstall entry
7. **Never** touches user projects in `%USERPROFILE%\Documents\Slate`

Internet is required during install.

## Build (this helper)

Requirements: Go 1.22+, Wails v2, Node 20+.

```powershell
wails build
powershell -File scripts\make-release-package.ps1
```

| Output | |
|--------|--|
| `dist/SlateSetup-windows-v*.zip` | User download package |
| `build/bin/SlateSetup.exe` | Installer binary |

## User package

- `SlateSetup.exe` — double-click
- `INSTALL.txt` / `README.txt` — how to install
- `NOTICE.txt` / `LICENSE.txt` — Sam’s app license + attribution

## License

- **Slate** (installed app): Apache-2.0, Sam Wasserman — LICENSE/NOTICE must be retained.
- **This Setup helper**: tooling to install his app on Windows; do not present it as official Wasserman software.
