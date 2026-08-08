# Build SlateSetup.exe and produce a user download package (exe + docs + zip).
# Usage:  powershell -File scripts/make-release-package.ps1
$ErrorActionPreference = "Stop"
$ver = "1.2.0"
$root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $root

Write-Host "Building SlateSetup..."
wails build
if ($LASTEXITCODE -ne 0) { throw "wails build failed" }

$pkgName = "SlateSetup-windows-v$ver"
$out = Join-Path $root "dist\$pkgName"
$stage = Join-Path $root "dist\package-staging"
if (Test-Path $out) { Remove-Item $out -Recurse -Force }
New-Item -ItemType Directory -Force -Path $out | Out-Null

Copy-Item (Join-Path $root "build\bin\SlateSetup.exe") (Join-Path $out "SlateSetup.exe") -Force
Copy-Item (Join-Path $stage "README.txt") (Join-Path $out "README.txt") -Force
Copy-Item (Join-Path $stage "INSTALL.txt") (Join-Path $out "INSTALL.txt") -Force
Copy-Item (Join-Path $stage "NOTICE.txt") (Join-Path $out "NOTICE.txt") -Force

$licCandidates = @(
  (Join-Path $root "..\slate-0.3.2\LICENSE"),
  (Join-Path $root "..\slate\LICENSE")
)
foreach ($lic in $licCandidates) {
  if (Test-Path $lic) {
    Copy-Item $lic (Join-Path $out "LICENSE.txt") -Force
    break
  }
}
if (-not (Test-Path (Join-Path $out "LICENSE.txt"))) {
  Set-Content (Join-Path $out "LICENSE.txt") "Apache License 2.0 - see https://www.apache.org/licenses/LICENSE-2.0"
}

$zip = Join-Path $root "dist\$pkgName.zip"
if (Test-Path $zip) { Remove-Item $zip -Force }
Compress-Archive -Path (Join-Path $out "*") -DestinationPath $zip -CompressionLevel Optimal

Write-Host ""
Write-Host "Release ready:"
Write-Host "  Folder: $out"
Write-Host "  Zip:    $zip"
Get-ChildItem $out | Format-Table Name, Length
Get-Item $zip | Format-List FullName, Length
