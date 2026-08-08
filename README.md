# Slate Setup (Windows)

Guided GUI installer (`SlateSetup.exe`) so people can get **Slate** running on Windows without using the terminal.

## Credits (important)

| Who | Role |
|-----|------|
| **[Sam Wasserman](https://wassermanproductions.com)** | **Creator of Slate** (the application), Apache-2.0 |
| **[ChrisToast89](https://github.com/ChrisToast89)** | Fork curator / Windows install tooling only — **not** the app author |

This repository is **community Windows Setup tooling**. It is **not** an official Wasserman Productions release and must not be described as Sam’s product.

- Official Slate app: https://github.com/wassermanproductions/slate  
- Windows-support app fork: https://github.com/ChrisToast89/slate (`windows-support` branch)

## What Setup does

1. Checks the PC (Windows version, disk, Node, ffmpeg, etc.)
2. Installs missing tools when possible (Node.js LTS, ffmpeg via winget)
3. Downloads Slate source from **ChrisToast89/slate@windows-support** (Windows-ready fork of Sam’s app)
4. Applies Windows brain fix if still needed
5. Builds the Electron app and installs to `%LOCALAPPDATA%\Programs\Slate`
6. Start Menu shortcut + uninstall entry; **never** deletes `%USERPROFILE%\Documents\Slate` projects
7. Optional Claude Code guidance (no API keys, no automated OAuth)

Internet is required during install.

## Build (developers)

Requirements: Go 1.22+, Wails v2, Node 20+.

```powershell
cd slate-windows-setup   # or slate-installer local folder
wails build
powershell -File scripts\make-release-package.ps1
```

| Artifact | Purpose |
|----------|---------|
| `dist/SlateSetup-windows-v*.zip` | **Share with users** |
| `build/bin/SlateSetup.exe` | Raw installer binary |

## User package contents

- `SlateSetup.exe` — double-click installer  
- `INSTALL.txt` / `README.txt` — support docs  
- `NOTICE.txt` / `LICENSE.txt` — Apache-2.0 + attribution  

## Related repos

| Repo | Purpose |
|------|---------|
| [wassermanproductions/slate](https://github.com/wassermanproductions/slate) | Official Slate (Sam) |
| [ChrisToast89/slate](https://github.com/ChrisToast89/slate) | Fork + `windows-support` branch |
| **This repo** | Windows Setup installer only |

## License

Slate application: Apache-2.0, **Sam Wasserman** — retain LICENSE/NOTICE.  
This Setup helper: community tooling for Windows; same attribution requirements for the app it installs.
