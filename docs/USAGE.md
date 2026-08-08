# Using Slate Setup (for non-technical users)

1. Double-click **`SlateSetup.exe`**.
2. On the **Home** screen you will see whether Slate is already installed.
3. Pick what you need:

| Button | What it does |
|--------|----------------|
| **Install / Repair** | Installs or repairs the app |
| **Audit tool** | Read-only check of your PC, install health, and GitHub updates |
| **Check for updates** | Compares your install to the GitHub repository |
| **Install update** | Appears only when a newer version is found |
| **Open my projects folder** | Opens `Documents\Slate` (never modified by Setup) |

4. If installing/updating: follow **Check PC → Install** and wait (often 10–20 minutes the first time).

## Your projects are protected

- Projects live in **`%USERPROFILE%\Documents\Slate`**
- Setup **only** replaces program files under **`%LOCALAPPDATA%\Programs\Slate`**
- Updates and uninstall **do not** delete your project JSON or media links

## After install

- Start Menu → **Slate**
- Uninstall: **Settings → Apps → Slate** (removes the app only, not Documents\Slate)

## Sign in for AI (Claude)

1. Optional: install Claude Code  
2. Run `claude auth login` in a terminal  
3. In Slate, set Brain to Claude Code  

## Problems?

Log: `%TEMP%\slate-install.log` — safe to re-run Setup anytime.
