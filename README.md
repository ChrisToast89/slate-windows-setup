# Slate Setup for Windows

A **helper installer** so Windows users can run **[Slate](https://github.com/wassermanproductions/slate)** without wrestling with Node, npm, and build steps.

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
