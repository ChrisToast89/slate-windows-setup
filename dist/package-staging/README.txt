================================================================================
  SLATE SETUP for Windows
  Helper installer for Sam Wasserman's Slate
================================================================================

WHO MADE WHAT
-------------
  SLATE (the application) was created by Sam Wasserman (Apache License 2.0).
    https://wassermanproductions.com
    https://github.com/wassermanproductions/slate

  THIS SETUP PROGRAM is only a Windows helper. It downloads and builds HIS app
  so Windows users can run it more easily. It is not written by Sam Wasserman
  and is not an official Wasserman Productions product.


WHAT YOU NEED
-------------
  - Windows 10 or Windows 11
  - Internet connection (required during install)
  - About 2 GB free disk space
  - 10-20 minutes the first time (downloads tools and builds the app)

You do not need Node, Git, or developer tools first. Setup handles that when it can.


HOW TO INSTALL
--------------
  1. Unzip this folder if you received a ZIP file.
  2. Double-click  SlateSetup.exe
  3. If Windows SmartScreen warns you, choose: More info -> Run anyway
  4. On the Home screen, click Install Slate (or Repair if needed).
  5. Click through Check this PC -> Continue -> Install Slate
  6. Wait until Finished. Leave the window open.
  7. Open Slate from the Start Menu.


WHAT SETUP DOES
---------------
  - Checks your computer
  - Installs free tools if missing (Node.js, ffmpeg) when possible
  - Downloads Slate from Sam's official GitHub repository
  - Applies a small Windows-only fix so the AI "brain" CLI works on Windows
  - Builds and installs the app for your user account
  - Creates a Start Menu shortcut

App install location (program files only):
  %LOCALAPPDATA%\Programs\Slate

Your projects (never deleted by Setup or updates):
  %USERPROFILE%\Documents\Slate


AUDIT AND UPDATES
-----------------
  - Audit tool: checks your PC and install without changing anything
  - Check for updates: compares to Sam's official GitHub repo
  - Install update: updates program files only; keeps your projects


AI BRAIN (OPTIONAL)
-------------------
  Slate does not use API keys. For AI help you can use Claude Code:

  1. Optional button in Setup: Install Claude Code
  2. Or: npm install -g @anthropic-ai/claude-code
  3. Sign in: claude auth login (opens your browser)
  4. In Slate, set Brain to Claude Code

  Or use a free local model (Ollama, LM Studio, etc.).


IF SOMETHING GOES WRONG
-----------------------
  1. Read the on-screen message.
  2. Log file: %TEMP%\slate-install.log
  3. Run SlateSetup.exe again (safe to re-run).
  4. Confirm internet and free disk space.


UNINSTALL
---------
  Windows Settings -> Apps -> Slate -> Uninstall
  Removes the app only. Projects in Documents\Slate are kept.


SUPPORT THE AUTHOR OF SLATE
---------------------------
  Sam Wasserman:
    https://ko-fi.com/samwasserman
    https://github.com/sponsors/wassermanproductions
    https://github.com/wassermanproductions/slate


PACKAGE CONTENTS
----------------
  SlateSetup.exe  - the installer (double-click this)
  README.txt      - this file
  INSTALL.txt     - short checklist
  NOTICE.txt      - attribution
  LICENSE.txt     - Apache License 2.0 (Slate)

Setup helper version: 1.2.0
Official Slate source: https://github.com/wassermanproductions/slate
================================================================================
