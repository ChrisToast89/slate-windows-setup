================================================================================
  SLATE SETUP for Windows
  Community installer for Slate — the prompt studio for AI filmmaking
================================================================================

ABOUT CREDITS (please read)
---------------------------
  Slate the application was created by Sam Wasserman (Apache License 2.0).
    https://wassermanproductions.com
    https://github.com/wassermanproductions/slate

  This Windows Setup package is NOT from Sam Wasserman.
  It is community Windows port / install tooling (fork curator: ChrisToast89).
  The maintainer only curates the Windows install experience — not the app.


WHAT YOU NEED
-------------
  • Windows 10 or Windows 11
  • Internet connection (required during install)
  • About 2 GB free disk space
  • 10–20 minutes the first time (downloads tools and builds the app)

You do NOT need to install Node, Git, or anything else first. Setup will
handle that for you when possible.


HOW TO INSTALL (simple)
-----------------------
  1. Unzip this folder if you received a ZIP file.
  2. Double-click  SlateSetup.exe
  3. If Windows SmartScreen warns you (unsigned app), choose
       More info  →  Run anyway
  4. On the Home screen, click  Install Slate  (or Repair if you already
     have it).
  5. Click through  Check this PC  →  Continue  →  Install Slate
  6. Wait until you see Finished. Leave the window open.
  7. Click  Open Slate  or find  Slate  in the Start Menu.


WHAT SETUP DOES FOR YOU
-----------------------
  • Checks your computer
  • Installs free tools if missing (Node.js, ffmpeg) when possible
  • Downloads Windows-ready Slate source (community fork of Sam's project)
  • Builds and installs; Windows CLI fix is included in that source
  • Builds and installs the app for your user account
  • Creates a Start Menu shortcut

App install location (program files only):
  %LOCALAPPDATA%\Programs\Slate

Your creative projects (never deleted by Setup or updates):
  %USERPROFILE%\Documents\Slate


AUDIT AND UPDATES
-----------------
  • Audit tool — checks your PC and install without changing anything
  • Check for updates — compares your install to the official GitHub repo
  • Install update — updates program files only; keeps your projects


AI BRAIN (OPTIONAL)
-------------------
  Slate does not use API keys. For AI help inside the app you can use
  Claude Code after install:

  1. Optional button in Setup: Install Claude Code
  2. Or later, in a terminal:  npm install -g @anthropic-ai/claude-code
  3. Sign in:  claude auth login  (opens your browser)
  4. In Slate, set Brain to Claude Code

  You can also use a free local model (Ollama, LM Studio, etc.) instead.


IF SOMETHING GOES WRONG
-----------------------
  1. Read the message on screen carefully.
  2. Open the log file:
       %TEMP%\slate-install.log
  3. Run SlateSetup.exe again — it is safe to re-run.
  4. Make sure you have internet and enough disk space.


UNINSTALL
---------
  Windows Settings → Apps → Slate → Uninstall

  This removes the app only. Your projects in Documents\Slate are kept.


SUPPORT THE ORIGINAL CREATOR
----------------------------
  Sam Wasserman (creator of Slate):
    Ko-fi:     https://ko-fi.com/samwasserman
    Sponsors:  https://github.com/sponsors/wassermanproductions
    Source:    https://github.com/wassermanproductions/slate


PACKAGE CONTENTS
----------------
  SlateSetup.exe     — the installer (double-click this)
  README.txt         — this file
  INSTALL.txt        — short one-page checklist
  NOTICE.txt         — attribution (original author + this package)
  LICENSE.txt        — Apache License 2.0

Setup package version: 1.1.2
Fork (Windows source): https://github.com/ChrisToast89/slate (branch windows-support)
================================================================================
