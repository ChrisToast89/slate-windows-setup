# Using Slate Setup (Windows)

This program installs **Sam Wasserman's Slate** (Electron/npm) on Windows. It is a helper only.

**Not Win-Slate** (the separate Go+Wails port). Product identity: [../AGENTS.md](../AGENTS.md).

## Install

1. Double-click `SlateSetup.exe`
2. If Windows blocks it: More info -> Run anyway
3. Click Install Slate
4. Wait for Finished (first run often 10-20 minutes with internet)
5. Open Slate from the Start Menu

## What stays protected

- Projects: `%USERPROFILE%\Documents\Slate` (never deleted by Setup)
- App files: `%LOCALAPPDATA%\Programs\Slate`

## Audit / updates

- Audit tool: read-only system and install check
- Check for updates: against Sam's official GitHub repo
- Install update: replaces program files only

## Credits

- Slate app: Sam Wasserman - https://github.com/wassermanproductions/slate
- This Setup helper: Windows install aid only (not by Sam)

## Log

`%TEMP%\slate-install.log`
